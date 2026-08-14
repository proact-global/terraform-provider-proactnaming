// Copyright (c) Proact
// SPDX-License-Identifier: MIT

package provider

import "testing"

// TestGetAzurermResourceType exercises the lookup that backs both the
// azurerm_resource_type attribute on proactnaming_resource_types and the whole
// proactnaming_azurerm_resources data source. It needs no live naming tool.
func TestGetAzurermResourceType(t *testing.T) {
	tests := []struct {
		name     string
		resource string
		property string
		want     string
	}{
		{
			name:     "exact match",
			resource: "storage/storageaccounts",
			want:     "azurerm_storage_account",
		},
		{
			name:     "lookup is case insensitive",
			resource: "Storage/StorageAccounts",
			want:     "azurerm_storage_account",
		},
		{
			name:     "several azurerm types share one ARM type",
			resource: "compute/virtualmachines",
			want:     "azurerm_linux_virtual_machine,azurerm_virtual_machine,azurerm_windows_virtual_machine",
		},
		{
			name:     "unknown resource yields an empty string",
			resource: "not/a-real-resource-type",
			want:     "",
		},
		{
			name:     "unmatched property falls back to the resource-only entry",
			resource: "resources/resourcegroups",
			property: "no-such-property",
			want:     "azurerm_resource_group",
		},
		{
			name:     "empty resource yields an empty string",
			resource: "",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getAzurermResourceType(tt.resource, tt.property); got != tt.want {
				t.Errorf("getAzurermResourceType(%q, %q) = %q, want %q",
					tt.resource, tt.property, got, tt.want)
			}
		})
	}
}
