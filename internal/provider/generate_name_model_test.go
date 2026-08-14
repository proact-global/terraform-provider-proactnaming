// Copyright (c) Proact
// SPDX-License-Identifier: MIT

package provider

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestBuildCustomComponents pins how the application attribute and the
// custom_component blocks are merged into the single map the API expects.
// It needs no live naming tool.
func TestBuildCustomComponents(t *testing.T) {
	tests := []struct {
		name  string
		model generateNameModel
		want  map[string]string
	}{
		{
			name:  "application becomes a custom component",
			model: generateNameModel{Application: types.StringValue("webapp")},
			want:  map[string]string{"application": "webapp"},
		},
		{
			name: "custom_component blocks are merged alongside application",
			model: generateNameModel{
				Application: types.StringValue("webapp"),
				CustomComponents: []customComponentModel{
					{Name: types.StringValue("subnet_tier"), Value: types.StringValue("app")},
					{Name: types.StringValue("subnet_instance"), Value: types.StringValue("01")},
				},
			},
			want: map[string]string{
				"application":     "webapp",
				"subnet_tier":     "app",
				"subnet_instance": "01",
			},
		},
		{
			name:  "a null application is omitted entirely",
			model: generateNameModel{Application: types.StringNull()},
			want:  map[string]string{},
		},
		{
			name:  "an unknown application is omitted entirely",
			model: generateNameModel{Application: types.StringUnknown()},
			want:  map[string]string{},
		},
		{
			name: "blocks with a null or unknown value are skipped",
			model: generateNameModel{
				Application: types.StringValue("webapp"),
				CustomComponents: []customComponentModel{
					{Name: types.StringValue("skipped_null"), Value: types.StringNull()},
					{Name: types.StringValue("skipped_unknown"), Value: types.StringUnknown()},
					{Name: types.StringValue("kept"), Value: types.StringValue("yes")},
				},
			},
			want: map[string]string{"application": "webapp", "kept": "yes"},
		},
		{
			// Documents current behaviour rather than endorsing it: a block named
			// "application" silently wins over the application attribute, because
			// the blocks are applied after it. Worth surfacing to users at some
			// point; pinned here so the precedence cannot change unnoticed.
			name: "an application block silently overrides the application attribute",
			model: generateNameModel{
				Application: types.StringValue("from-attribute"),
				CustomComponents: []customComponentModel{
					{Name: types.StringValue("application"), Value: types.StringValue("from-block")},
				},
			},
			want: map[string]string{"application": "from-block"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.model.buildCustomComponents()
			if !reflect.DeepEqual(map[string]string(got), tt.want) {
				t.Errorf("buildCustomComponents() = %v, want %v", got, tt.want)
			}
		})
	}
}
