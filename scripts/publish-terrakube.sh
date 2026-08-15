#!/usr/bin/env bash
#
# Publish a released provider version to the Terrakube private registry.
#
# Terrakube stores registry metadata but not the binaries themselves, so each
# platform is registered by pointing at a URL. This script points at the GitHub
# release for the tag, which already carries the zips, SHA256SUMS and the
# detached signature that GoReleaser produced -- so nothing needs uploading
# anywhere first.
#
# Registering asciiArmor with Terrakube is what lets Terraform verify the
# signature, which is the reason the filesystem-mirror workaround in
# docs/USAGE.md existed. Published this way, that workaround is unnecessary.
#
# Usage:
#   scripts/publish-terrakube.sh [tag]            # dry run, prints what it would do
#   scripts/publish-terrakube.sh [tag] --publish  # actually publishes
#
# Tag defaults to the most recent v* tag.
#
# Required environment:
#   TERRAKUBE_TOKEN           Bearer token for the Terrakube API. Obtain with
#                             `terraform login registry.eng.proactmcs.eu`, which
#                             runs the dex/Entra ID flow and stores a token in
#                             ~/.terraform.d/credentials.tfrc.json, or use a
#                             Terrakube personal access token.
#   TERRAKUBE_GPG_KEY_ID      Long key id of the signing key, e.g. 34365D9472D7468F.
#   TERRAKUBE_GPG_PUBLIC_KEY  ASCII-armored public key, matching the private key
#                             in the GPG_PRIVATE_KEY action secret. Export with:
#                               gpg --armor --export "$TERRAKUBE_GPG_KEY_ID"
#
# Optional environment:
#   TERRAKUBE_API             Default https://terrakube.eng.proactmcs.eu
#   TERRAKUBE_ORG_ID          Default is the proact organization id below.
#   GITHUB_REPO               Default proact-global/terraform-provider-proactnaming
#   PROVIDER_NAME             Default proactnaming
#   PROVIDER_PROTOCOL         Default 6.0, matching terraform-registry-manifest.json.

set -euo pipefail

TERRAKUBE_API="${TERRAKUBE_API:-https://terrakube.eng.proactmcs.eu}"
TERRAKUBE_ORG_ID="${TERRAKUBE_ORG_ID:-e35fcdcd-8011-4b42-aefb-065ce6d12912}"
GITHUB_REPO="${GITHUB_REPO:-proact-global/terraform-provider-proactnaming}"
PROVIDER_NAME="${PROVIDER_NAME:-proactnaming}"
PROVIDER_PROTOCOL="${PROVIDER_PROTOCOL:-6.0}"

JSONAPI="Content-Type: application/vnd.api+json"

TAG="${1:-}"
PUBLISH=false
for arg in "$@"; do
  [ "$arg" = "--publish" ] && PUBLISH=true
done
[ "${TAG:-}" = "--publish" ] && TAG=""

die() { echo "error: $*" >&2; exit 1; }
step() { echo; echo "==> $*"; }

command -v curl    >/dev/null || die "curl is required"
command -v python3 >/dev/null || die "python3 is required"

# ---------------------------------------------------------------- preflight --
for v in TERRAKUBE_TOKEN TERRAKUBE_GPG_KEY_ID TERRAKUBE_GPG_PUBLIC_KEY; do
  if [ -z "${!v:-}" ]; then
    $PUBLISH && die "$v is not set (required to publish)"
    echo "warning: $v is not set -- dry run will continue, publishing would fail" >&2
  fi
done

if [ -z "$TAG" ]; then
  TAG=$(git tag -l 'v*' --sort=v:refname | tail -1)
  [ -n "$TAG" ] || die "no v* tag found and none given"
fi
VERSION="${TAG#v}"

echo "provider     : $PROVIDER_NAME"
echo "tag/version  : $TAG -> $VERSION"
echo "protocol     : $PROVIDER_PROTOCOL"
echo "terrakube    : $TERRAKUBE_API (org $TERRAKUBE_ORG_ID)"
echo "binaries from: github.com/$GITHUB_REPO releases"
$PUBLISH || echo "MODE         : dry run (pass --publish to apply)"

# ------------------------------------------------------- github release data --
step "Reading GitHub release $TAG"

gh_api() {
  local url="$1"
  if [ -n "${GITHUB_TOKEN:-}" ]; then
    curl -sfL -H "Authorization: Bearer $GITHUB_TOKEN" "$url"
  else
    curl -sfL "$url"
  fi
}

RELEASE_JSON=$(gh_api "https://api.github.com/repos/$GITHUB_REPO/releases/tags/$TAG") \
  || die "no GitHub release found for $TAG"

# The zips, SHA256SUMS and its signature are all assets of that release.
read -r SHASUMS_URL SIG_URL < <(python3 -c "
import json,sys
assets={a['name']: a['browser_download_url'] for a in json.loads(sys.argv[1])['assets']}
sums=[u for n,u in assets.items() if n.endswith('SHA256SUMS')]
sig =[u for n,u in assets.items() if n.endswith('SHA256SUMS.sig')]
if not sums: sys.exit('release has no SHA256SUMS asset')
if not sig:  sys.exit('release has no SHA256SUMS.sig asset')
print(sums[0], sig[0])
" "$RELEASE_JSON") || die "could not locate checksum assets"

echo "  SHA256SUMS     : $SHASUMS_URL"
echo "  SHA256SUMS.sig : $SIG_URL"

step "Fetching checksums"
SHASUMS=$(curl -sfL "$SHASUMS_URL") || die "could not download SHA256SUMS"
echo "  $(echo "$SHASUMS" | grep -c . ) entries"

# Build one implementation payload per platform zip, pairing each asset with its
# checksum. A zip without a matching SHA256SUMS line is a broken release and is
# treated as fatal rather than skipped.
PLATFORMS_JSON=$(python3 - "$RELEASE_JSON" "$SHASUMS" "$VERSION" <<'PY'
import json, sys, re

release, shasums_text, version = sys.argv[1], sys.argv[2], sys.argv[3]
assets = {a['name']: a['browser_download_url'] for a in json.loads(release)['assets']}

sums = {}
for line in shasums_text.splitlines():
    parts = line.split()
    if len(parts) == 2:
        sums[parts[1].lstrip('*')] = parts[0]

out, missing = [], []
pattern = re.compile(rf'_{re.escape(version)}_(?P<os>[a-z0-9]+)_(?P<arch>[a-z0-9]+)\.zip$')
for name, url in sorted(assets.items()):
    if not name.endswith('.zip'):
        continue
    m = pattern.search(name)
    if not m:
        continue
    if name not in sums:
        missing.append(name)
        continue
    out.append({
        'os': m.group('os'), 'arch': m.group('arch'),
        'filename': name, 'downloadUrl': url, 'shasum': sums[name],
    })

if missing:
    sys.exit('assets missing from SHA256SUMS: ' + ', '.join(missing))
if not out:
    sys.exit('no platform zips matched version ' + version)
print(json.dumps(out))
PY
) || die "could not build platform list"

echo "  $(python3 -c "import json,sys; print(len(json.loads(sys.argv[1])))" "$PLATFORMS_JSON") platforms matched"

# ------------------------------------------------------------- terrakube api --
tk() {
  local method="$1" path="$2" body="${3:-}"
  local args=(-sS -X "$method" -H "Authorization: Bearer ${TERRAKUBE_TOKEN:-}" -H "$JSONAPI"
              -H "Accept: application/vnd.api+json" -w '\n%{http_code}')
  [ -n "$body" ] && args+=(-d "$body")
  curl "${args[@]}" "$TERRAKUBE_API$path"
}

tk_ok() {  # prints body, fails on non-2xx
  local out code
  out=$(tk "$@") || return 1
  code=$(printf '%s' "$out" | tail -1)
  body=$(printf '%s' "$out" | sed '$d')
  case "$code" in
    2*) printf '%s' "$body"; return 0 ;;
    *)  echo "  HTTP $code: $body" >&2; return 1 ;;
  esac
}

ORG_PATH="/api/v1/organization/$TERRAKUBE_ORG_ID"

step "Locating provider '$PROVIDER_NAME'"
if $PUBLISH; then
  EXISTING=$(tk_ok GET "$ORG_PATH/provider") || die "could not list providers (check TERRAKUBE_TOKEN)"
  PROVIDER_ID=$(python3 -c "
import json,sys
try: d=json.loads(sys.argv[1])
except Exception: sys.exit(0)
for p in d.get('data',[]) or []:
    if p.get('attributes',{}).get('name')==sys.argv[2]: print(p['id']); break
" "$EXISTING" "$PROVIDER_NAME")

  if [ -z "$PROVIDER_ID" ]; then
    echo "  not found, creating"
    CREATED=$(tk_ok POST "$ORG_PATH/provider" "$(python3 -c "
import json,sys
print(json.dumps({'data':{'type':'provider','attributes':{
  'name': sys.argv[1],
  'description': 'Generates standardized Azure resource names using the Azure Naming Tool.'}}}))
" "$PROVIDER_NAME")") || die "could not create provider"
    PROVIDER_ID=$(python3 -c "import json,sys; print(json.loads(sys.argv[1])['data']['id'])" "$CREATED")
  fi
  echo "  provider id: $PROVIDER_ID"
else
  PROVIDER_ID='<resolved-at-publish-time>'
  echo "  dry run: would GET $ORG_PATH/provider and create it if absent"
fi

step "Creating version $VERSION"
VERSION_BODY=$(python3 -c "
import json,sys
print(json.dumps({'data':{'type':'version','attributes':{
  'versionNumber': sys.argv[1], 'protocols': sys.argv[2]}}}))
" "$VERSION" "$PROVIDER_PROTOCOL")

if $PUBLISH; then
  CREATED=$(tk_ok POST "$ORG_PATH/provider/$PROVIDER_ID/version" "$VERSION_BODY") \
    || die "could not create version $VERSION (already published?)"
  VERSION_ID=$(python3 -c "import json,sys; print(json.loads(sys.argv[1])['data']['id'])" "$CREATED")
  echo "  version id: $VERSION_ID"
else
  VERSION_ID='<resolved-at-publish-time>'
  echo "  dry run: POST $ORG_PATH/provider/<id>/version"
  echo "  $VERSION_BODY"
fi

step "Registering platform implementations"
python3 - "$PLATFORMS_JSON" "$SHASUMS_URL" "$SIG_URL" \
         "${TERRAKUBE_GPG_KEY_ID:-<unset>}" "${TERRAKUBE_GPG_PUBLIC_KEY:-<unset>}" <<'PY' > /tmp/tk_impls.txt
import json, sys
plats, sums_url, sig_url, key_id, armor = sys.argv[1:6]
for p in json.loads(plats):
    print(json.dumps({'data': {'type': 'implementation', 'attributes': {
        'os': p['os'], 'arch': p['arch'], 'filename': p['filename'],
        'downloadUrl': p['downloadUrl'],
        'shasumsUrl': sums_url, 'shasumsSignatureUrl': sig_url,
        'shasum': p['shasum'], 'keyId': key_id, 'asciiArmor': armor,
    }}}))
PY

COUNT=0
while IFS= read -r payload; do
  DESC=$(python3 -c "
import json,sys
a=json.loads(sys.argv[1])['data']['attributes']
print(f\"{a['os']}/{a['arch']}  {a['shasum'][:12]}...\")" "$payload")
  if $PUBLISH; then
    tk_ok POST "$ORG_PATH/provider/$PROVIDER_ID/version/$VERSION_ID/implementation" "$payload" >/dev/null \
      && echo "  ok   $DESC" || echo "  FAIL $DESC"
  else
    echo "  would register $DESC"
  fi
  COUNT=$((COUNT+1))
done < /tmp/tk_impls.txt
rm -f /tmp/tk_impls.txt

step "Done"
echo "  $COUNT platforms for $PROVIDER_NAME $VERSION"
if $PUBLISH; then
  echo
  echo "  Consume it with:"
  echo "    source  = \"registry.eng.proactmcs.eu/proact/$PROVIDER_NAME\""
  echo "    version = \"~> ${VERSION%.*}\""
else
  echo "  Dry run only. Re-run with --publish to apply."
fi
