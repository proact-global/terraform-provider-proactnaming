// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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

func testAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for example assertions.
	// about the appropriate environment variables being set are common to see in a pre-check.
	// function.

	// Check for required environment variables for acceptance tests.
	// These can be set for testing without exposing credentials in code.
	if v := os.Getenv("PROACTNAMING_HOST"); v == "" {
		t.Skip("PROACTNAMING_HOST must be set for acceptance tests")
	}
	if v := os.Getenv("PROACTNAMING_APIKEY"); v == "" {
		t.Skip("PROACTNAMING_APIKEY must be set for acceptance tests")
	}
}
