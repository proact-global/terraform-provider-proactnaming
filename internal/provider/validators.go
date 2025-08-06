package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// stringLengthValidator validates that a string attribute has a length within the specified range.
type stringLengthValidator struct {
	min int
	max int
}

func (v stringLengthValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("string length must be between %d and %d characters", v.min, v.max)
}

func (v stringLengthValidator) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("string length must be between %d and %d characters", v.min, v.max)
}

func (v stringLengthValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.ValueString()
	if len(value) < v.min || len(value) > v.max {
		response.Diagnostics.AddAttributeError(
			request.Path,
			"Invalid String Length",
			fmt.Sprintf("Expected string length to be between %d and %d characters, got %d characters.", v.min, v.max, len(value)),
		)
	}
}

// StringLength returns a validator which ensures that any configured attribute value
// has a length between the given minimum and maximum values.
func StringLength(min, max int) validator.String {
	return stringLengthValidator{
		min: min,
		max: max,
	}
}

// stringPatternValidator validates that a string attribute matches a specific pattern.
type stringPatternValidator struct {
	pattern *regexp.Regexp
	message string
}

func (v stringPatternValidator) Description(ctx context.Context) string {
	return v.message
}

func (v stringPatternValidator) MarkdownDescription(ctx context.Context) string {
	return v.message
}

func (v stringPatternValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.ValueString()
	if !v.pattern.MatchString(value) {
		response.Diagnostics.AddAttributeError(
			request.Path,
			"Invalid String Pattern",
			fmt.Sprintf("Expected string to match pattern %s. %s", v.pattern.String(), v.message),
		)
	}
}

// StringPattern returns a validator which ensures that any configured attribute value
// matches the given regular expression pattern.
func StringPattern(pattern string, message string) validator.String {
	return stringPatternValidator{
		pattern: regexp.MustCompile(pattern),
		message: message,
	}
}

// stringNotEmptyValidator validates that a string attribute is not empty.
type stringNotEmptyValidator struct{}

func (v stringNotEmptyValidator) Description(ctx context.Context) string {
	return "string must not be empty"
}

func (v stringNotEmptyValidator) MarkdownDescription(ctx context.Context) string {
	return "string must not be empty"
}

func (v stringNotEmptyValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.ValueString()
	if strings.TrimSpace(value) == "" {
		response.Diagnostics.AddAttributeError(
			request.Path,
			"Invalid String Value",
			"String must not be empty or contain only whitespace.",
		)
	}
}

// StringNotEmpty returns a validator which ensures that any configured attribute value
// is not an empty string or contains only whitespace.
func StringNotEmpty() validator.String {
	return stringNotEmptyValidator{}
}
