#!/usr/bin/env bash
# Copyright (c) Proact
# SPDX-License-Identifier: MIT

#
# Build the provider and run Terraform against it, without publishing a release
# or installing anything from the registry.
#
# Terraform is pointed at the freshly built binary through a development
# override, so each run uses the working tree rather than a released version.
#
# Usage:
#   scripts/local-dev.sh                 # terraform plan
#   scripts/local-dev.sh apply           # plan and apply
#   scripts/local-dev.sh destroy         # remove what was created
#   scripts/local-dev.sh plan -var=predict_names_locally=true
#
# Required environment, the same variables the provider reads normally:
#   PROACTNAMING_HOST            base URL, including https:// and no trailing slash
#   PROACTNAMING_APIKEY          API key
#   PROACTNAMING_ADMIN_PASSWORD  admin password
#
# The admin password is needed for planning as well as destroying, because the
# preview entry created while planning is removed through the Admin API. It is
# not needed when predict_names_locally is set.

set -euo pipefail

cd "$(dirname "$0")/.."
repo=$PWD
example=$repo/examples/local-dev
workdir=${LOCAL_DEV_DIR:-$repo/.local-dev}

command -v terraform >/dev/null || { echo "error: terraform is not installed" >&2; exit 1; }
command -v go >/dev/null || { echo "error: go is not installed" >&2; exit 1; }

command=${1:-plan}
[ $# -gt 0 ] && shift

missing=()
for v in PROACTNAMING_HOST PROACTNAMING_APIKEY; do
  [ -z "${!v:-}" ] && missing+=("$v")
done
if [ ${#missing[@]} -gt 0 ]; then
  echo "error: ${missing[*]} not set." >&2
  echo >&2
  echo "  export PROACTNAMING_HOST=\"https://your-naming-tool.azurewebsites.net\"" >&2
  echo "  export PROACTNAMING_APIKEY=\"...\"" >&2
  echo "  export PROACTNAMING_ADMIN_PASSWORD=\"...\"" >&2
  exit 1
fi

if [ -z "${PROACTNAMING_ADMIN_PASSWORD:-}" ]; then
  echo "note: PROACTNAMING_ADMIN_PASSWORD is not set. Planning removes the preview" >&2
  echo "      entry it creates through the Admin API, so it will warn that the entry" >&2
  echo "      was left behind. Pass -var=predict_names_locally=true to avoid needing it." >&2
  echo >&2
fi

# The override points at a directory, and Terraform looks inside it for a
# binary named after the provider.
bindir=$workdir/bin
mkdir -p "$bindir" "$workdir/tf"

echo "==> building the provider from the working tree"
go build -o "$bindir/terraform-provider-proactnaming" .

cat > "$workdir/terraformrc" <<EOF
provider_installation {
  dev_overrides {
    "proact-global/proactnaming" = "$bindir"
  }
  direct {}
}
EOF

cp "$example"/*.tf "$workdir/tf/"

echo "==> terraform $command"
echo
cd "$workdir/tf"
export TF_CLI_CONFIG_FILE="$workdir/terraformrc"

case "$command" in
  apply)   terraform apply -auto-approve "$@" ;;
  destroy) terraform destroy -auto-approve "$@" ;;
  *)       terraform "$command" "$@" ;;
esac

echo
echo "==> state and config are under $workdir"
echo "    Terraform reports the development override as a warning on every run;"
echo "    that is expected and means the working tree is in use."
