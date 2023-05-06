package validators

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// ensure implementation satisfied expected interfaces.
var _ validator.List = enumStringList{}

// EnumStringListValuesAre returns a validator which ensurses that any values given in the list are one of
// the given enumerated types.
func EnumStringListValuesAre(ignoreCase bool, enumValues ...string) validator.List {
	return enumStringList{
		values:     enumValues,
		ignoreCase: ignoreCase,
	}
}

// enumStringList holds details about the enumerated string list validator.
type enumStringList struct {
	// values holds the list of valid values for the enumeration.
	values []string

	// ignoreCase determines whether or not the values are case-sensitive.
	ignoreCase bool
}

// Description returns a plain text description of the validator's behavior, suitable for a practitioner to
// understand its impact.
func (v enumStringList) Description(ctx context.Context) string {
	return "checks that each value in the list matches one of the valid enumerated values."
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior, suitable for a
// practitioner to understand its impact.
func (v enumStringList) MarkdownDescription(ctx context.Context) string {
	return "checks that each value in the list matches one of the valid enumerated values."
}

// Validate runs the main validation logic of the validator, reading configuration data out of `req` and
// updating `resp` with diagnostics.
func (v enumStringList) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	_, ok := req.ConfigValue.ElementType(ctx).(basetypes.StringTypable)
	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Validator for Element Type",
			"While performing schema-based validation, an unexpected error occurred. "+
				"The attribute declares a String values validator, however its values do not implement types.StringType or "+
				"the types.StringTypable interface for custom String types. "+
				"Use the appropriate values validator that matches the element type. "+
				"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
				fmt.Sprintf("Path: %s\n", req.Path.String())+
				fmt.Sprintf("Element Type: %T\n", req.ConfigValue.ElementType(ctx)),
		)
		return
	}

	for i, element := range req.ConfigValue.Elements() {
		elementPath := req.Path.AtListIndex(i)
		elementValue := element.String()
		for _, val := range v.values {
			if v.ignoreCase {
				if strings.EqualFold(elementValue, val) {
					return
				}
			} else if elementValue == val {
				return
			}
		}
		resp.Diagnostics.AddAttributeError(
			elementPath,
			"Invalid Value Used",
			fmt.Sprintf("Value must be one of: %s", strings.Join(v.values, ", ")),
		)
	}
}
