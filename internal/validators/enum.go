package validators

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// ensure implementation satisfied expected interfaces.
var _ validator.String = enumString{}

// EnumStringValueOneOf returns a validator which ensurses that the values given is one of
// the given enumerated types.
func EnumStringValueOneOf(ignoreCase bool, enumValues ...string) validator.String {
	return enumString{
		values:     enumValues,
		ignoreCase: ignoreCase,
	}
}

// enumString holds details about the enumerated string list validator.
type enumString struct {
	// values holds the list of valid values for the enumeration.
	values []string

	// ignoreCase determines whether or not the values are case-sensitive.
	ignoreCase bool
}

// Description returns a plain text description of the validator's behavior, suitable for a practitioner to
// understand its impact.
func (v enumString) Description(ctx context.Context) string {
	return "checks that the value given matches one of the valid enumerated values."
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior, suitable for a
// practitioner to understand its impact.
func (v enumString) MarkdownDescription(ctx context.Context) string {
	return "checks that the value given matches one of the valid enumerated values."
}

// Validate runs the main validation logic of the validator, reading configuration data out of `req` and
// updating `resp` with diagnostics.
func (v enumString) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	value := req.ConfigValue.ValueString()
	for _, val := range v.values {
		if v.ignoreCase {
			if strings.EqualFold(value, val) {
				return
			}
		} else if value == val {
			return
		}
	}
	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Invalid Value Used",
		fmt.Sprintf("Value must be one of: %s", strings.Join(v.values, ", ")),
	)
}
