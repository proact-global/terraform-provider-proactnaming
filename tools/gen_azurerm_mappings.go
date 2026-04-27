// Copyright (c) HashiCorp, Inc.

//go:build ignore

// gen_azurerm_mappings.go generates internal/provider/azurerm_mappings.go by
// fetching the aztft map.json (https://github.com/magodo/aztft) and inverting
// the azurerm-resource-name → ARM-type mapping into the ARM-path → azurerm-names
// format used by the getAzurermResourceType helper.
//
// Run from the repository root:
//
//	go run tools/gen_azurerm_mappings.go
//
// Or via make:
//
//	make generate-mappings
//
// The -out flag overrides the default output path, useful when invoked from
// subdirectories (e.g. via go:generate in tools/):
//
//	go run gen_azurerm_mappings.go -out ../internal/provider/azurerm_mappings.go
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/format"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

const aztftMapURL = "https://raw.githubusercontent.com/magodo/aztft/main/internal/resmap/map.json"

// ManagementPlane holds the ARM provider and resource type hierarchy from aztft.
type ManagementPlane struct {
	Provider string   `json:"provider"`
	Types    []string `json:"types"`
}

// ResourceEntry represents one entry in aztft's map.json.
type ResourceEntry struct {
	IsRemoved       bool             `json:"is_removed,omitempty"`
	ManagementPlane *ManagementPlane `json:"management_plane,omitempty"`
}

func main() {
	out := flag.String("out", "internal/provider/azurerm_mappings.go", "output file path")
	flag.Parse()

	fmt.Printf("Fetching %s …\n", aztftMapURL)

	data, err := fetchMap()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	fmt.Printf("Fetched %d entries.\n", len(data))

	reverseMap, groupDisplay, sortedKeys := buildReverseMap(data)

	fmt.Printf("Generated %d unique ARM path keys.\n", len(reverseMap))

	src, err := renderFile(reverseMap, groupDisplay, sortedKeys)
	if err != nil {
		fmt.Fprintln(os.Stderr, "render error:", err)
		os.Exit(1)
	}

	if err := os.WriteFile(*out, src, 0644); err != nil {
		fmt.Fprintln(os.Stderr, "write error:", err)
		os.Exit(1)
	}

	fmt.Printf("Wrote %s\n", *out)
}

func fetchMap() (map[string]ResourceEntry, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Get(aztftMapURL)
	if err != nil {
		return nil, fmt.Errorf("fetch failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status: %s", resp.Status)
	}

	var data map[string]ResourceEntry
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("JSON decode failed: %w", err)
	}

	return data, nil
}

// buildReverseMap inverts the aztft map.
//
// Returns:
//   - reverseMap:   ARM path key → sorted list of azurerm resource names
//   - groupDisplay: lowercase namespace → display name (original provider casing minus "Microsoft.")
//   - sortedKeys:   all ARM path keys, sorted
func buildReverseMap(data map[string]ResourceEntry) (
	reverseMap map[string][]string,
	groupDisplay map[string]string,
	sortedKeys []string,
) {
	reverseMap = make(map[string][]string)
	groupDisplay = make(map[string]string)

	for azurermName, entry := range data {
		if entry.IsRemoved {
			continue
		}
		if entry.ManagementPlane == nil || len(entry.ManagementPlane.Types) == 0 {
			continue
		}

		mp := entry.ManagementPlane

		// Skip placeholder or templated providers (e.g. "{ResourceProviderName}").
		if strings.ContainsAny(mp.Provider, "{}") || mp.Provider == "" {
			continue
		}

		key := buildKey(mp.Provider, mp.Types)
		ns := strings.SplitN(key, "/", 2)[0]

		reverseMap[key] = append(reverseMap[key], azurermName)

		// Record the display name for this namespace group (first occurrence wins).
		if _, exists := groupDisplay[ns]; !exists {
			display := mp.Provider
			if strings.HasPrefix(strings.ToLower(display), "microsoft.") {
				display = display[len("Microsoft."):]
			}
			groupDisplay[ns] = display
		}
	}

	// Sort values within each key for determinism.
	for key := range reverseMap {
		sort.Strings(reverseMap[key])
	}

	// Collect and sort all keys.
	sortedKeys = make([]string, 0, len(reverseMap))
	for k := range reverseMap {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	return
}

// buildKey derives the lowercase ARM path key from a provider name and types slice.
//
// Examples:
//
//	("Microsoft.Compute",        ["virtualMachines"])             → "compute/virtualmachines"
//	("Microsoft.EventHub",       ["namespaces", "eventhubs"])     → "eventhub/namespaces/eventhubs"
//	("Microsoft.ApiManagement",  ["service"])                     → "apimanagement/service"
//	("Nginx.NginxPlus",          ["nginxDeployments"])            → "nginx.nginxplus/nginxdeployments"
//	("Microsoft.Resources",      ["subscriptions", "resourceGroups"]) → "resources/resourcegroups"
//
// Note: a leading "subscriptions" type is stripped because it is a subscription-scope
// qualifier in the ARM REST path that the Azure Naming Tool does not include in its keys.
func buildKey(provider string, types []string) string {
	ns := provider
	if strings.HasPrefix(strings.ToLower(ns), "microsoft.") {
		ns = ns[len("Microsoft."):]
	}

	filtered := types
	if len(filtered) > 1 && strings.EqualFold(filtered[0], "subscriptions") {
		filtered = filtered[1:]
	}

	parts := make([]string, 0, 1+len(filtered))
	parts = append(parts, ns)
	parts = append(parts, filtered...)

	return strings.ToLower(strings.Join(parts, "/"))
}

// renderFile produces a gofmt-formatted Go source file containing the
// generated azurermResourceTypeMap and the static getAzurermResourceType func.
func renderFile(
	reverseMap map[string][]string,
	groupDisplay map[string]string,
	sortedKeys []string,
) ([]byte, error) {
	var b bytes.Buffer

	fmt.Fprintf(&b, "// Code generated by tools/gen_azurerm_mappings.go; DO NOT EDIT.\n")
	fmt.Fprintf(&b, "// Source: %s\n", aztftMapURL)
	fmt.Fprintf(&b, "// Run 'make generate-mappings' to regenerate.\n")
	fmt.Fprintf(&b, "\n")
	fmt.Fprintf(&b, "package provider\n")
	fmt.Fprintf(&b, "\n")
	fmt.Fprintf(&b, "import (\n\t\"strings\"\n)\n")
	fmt.Fprintf(&b, "\n")

	// Map variable doc comment.
	fmt.Fprintf(&b, "// azurermResourceTypeMap maps lowercase AzureNamingTool resource paths to the\n")
	fmt.Fprintf(&b, "// corresponding azurerm Terraform provider resource type name(s).\n")
	fmt.Fprintf(&b, "//\n")
	fmt.Fprintf(&b, "// Keys use the format:\n")
	fmt.Fprintf(&b, "//   - \"namespace/resourcetype\"           (lowercase, no \"Microsoft.\" prefix)\n")
	fmt.Fprintf(&b, "//   - \"namespace/resourcetype/subtype\"  (for sub-resources)\n")
	fmt.Fprintf(&b, "//   - \"namespace/resourcetype|property\" (manual property-specific overrides)\n")
	fmt.Fprintf(&b, "//\n")
	fmt.Fprintf(&b, "// Values are sorted slices of azurerm resource type names. Property-specific\n")
	fmt.Fprintf(&b, "// keys (with the \"|property\" suffix) are NOT generated automatically — add them\n")
	fmt.Fprintf(&b, "// manually and they will be preserved across regeneration only if managed separately.\n")
	fmt.Fprintf(&b, "//\n")
	fmt.Fprintf(&b, "// Generated from: %s\n", aztftMapURL)
	fmt.Fprintf(&b, "var azurermResourceTypeMap = map[string][]string{\n")

	currentNS := ""
	for _, key := range sortedKeys {
		ns := strings.SplitN(key, "/", 2)[0]

		if ns != currentNS {
			if currentNS != "" {
				fmt.Fprintf(&b, "\n")
			}

			display := groupDisplay[ns]
			if display == "" {
				display = ns
			}

			fmt.Fprintf(&b, "\t// %s\n", display)
			currentNS = ns
		}

		values := reverseMap[key]
		if len(values) == 1 {
			fmt.Fprintf(&b, "\t%q: {%q},\n", key, values[0])
		} else {
			fmt.Fprintf(&b, "\t%q: {\n", key)
			for _, v := range values {
				fmt.Fprintf(&b, "\t\t%q,\n", v)
			}
			fmt.Fprintf(&b, "\t},\n")
		}
	}

	fmt.Fprintf(&b, "}\n")
	fmt.Fprintf(&b, "\n")

	// Lookup function — identical logic to the hand-written original.
	fmt.Fprintf(&b, "// getAzurermResourceType returns the azurerm Terraform provider resource type name(s)\n")
	fmt.Fprintf(&b, "// for a given AzureNamingTool resource path and optional property value.\n")
	fmt.Fprintf(&b, "// Returns a comma-separated string of matching resource types, or an empty string if unknown.\n")
	fmt.Fprintf(&b, "func getAzurermResourceType(resource, property string) string {\n")
	fmt.Fprintf(&b, "\tkey := strings.ToLower(resource)\n")
	fmt.Fprintf(&b, "\n")
	fmt.Fprintf(&b, "\t// Try property-specific match first.\n")
	fmt.Fprintf(&b, "\tif property != \"\" {\n")
	fmt.Fprintf(&b, "\t\tpropKey := key + \"|\" + strings.ToLower(property)\n")
	fmt.Fprintf(&b, "\t\tif matches, ok := azurermResourceTypeMap[propKey]; ok {\n")
	fmt.Fprintf(&b, "\t\t\treturn strings.Join(matches, \",\")\n")
	fmt.Fprintf(&b, "\t\t}\n")
	fmt.Fprintf(&b, "\t}\n")
	fmt.Fprintf(&b, "\n")
	fmt.Fprintf(&b, "\t// Fall back to resource-only match.\n")
	fmt.Fprintf(&b, "\tif matches, ok := azurermResourceTypeMap[key]; ok {\n")
	fmt.Fprintf(&b, "\t\treturn strings.Join(matches, \",\")\n")
	fmt.Fprintf(&b, "\t}\n")
	fmt.Fprintf(&b, "\n")
	fmt.Fprintf(&b, "\treturn \"\"\n")
	fmt.Fprintf(&b, "}\n")

	return format.Source(b.Bytes())
}
