// Copyright (c) Proact
// SPDX-License-Identifier: MIT

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories is used to instantiate a provider during acceptance testing.
// The factory function is called for each Terraform CLI command to create a provider.
// server that the CLI can connect to and interact with.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"proactnaming": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck ensures the credentials an acceptance test needs are present.
//
// Locally a missing credential skips the test, so `go test ./...` stays useful
// without a naming tool to hand. In CI it fails instead: a workflow that opts
// into acceptance testing but has no credentials configured would otherwise skip
// every test and report a green build while exercising nothing.
func testAccPreCheck(t *testing.T) {
	t.Helper()
	requireAccEnv(t, "PROACTNAMING_HOST")
	requireAccEnv(t, "PROACTNAMING_APIKEY")
}

// testAccPreCheckWithAdmin also requires the admin password, needed by tests
// that exercise the Admin API delete path.
func testAccPreCheckWithAdmin(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)
	requireAccEnv(t, "PROACTNAMING_ADMIN_PASSWORD")
}

// requireAccEnv skips or fails, depending on whether this is a CI run.
func requireAccEnv(t *testing.T, key string) {
	t.Helper()
	if os.Getenv(key) != "" {
		return
	}
	msg := key + " must be set for acceptance tests"
	if os.Getenv("CI") != "" {
		// Forcing a failure here is the point: silently skipping in CI is
		// indistinguishable from passing.
		t.Fatal(msg + " (set in CI, where skipping would report a false pass)")
	}
	t.Skip(msg)
}
