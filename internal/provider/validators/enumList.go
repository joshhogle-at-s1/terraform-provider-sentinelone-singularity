package validators

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/joshhogle-at-s1/terraform-provider-sentinelone-singularity/internal/plugin"
)

// ensure implementation satisfied expected interfaces
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
	return "checks that each value in the list matches one of the valid enumerated values"
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior, suitable for a
// practitioner to understand its impact.
func (v enumStringList) MarkdownDescription(ctx context.Context) string {
	return "checks that each value in the list matches one of the valid enumerated values"
}

// Validate runs the main validation logic of the validator, reading configuration data out of `req` and
// updating `resp` with diagnostics.
func (v enumStringList) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	_, ok := req.ConfigValue.ElementType(ctx).(basetypes.StringTypable)
	if !ok {
		// this should *never* happen - but we want to be sure
		msg := fmt.Sprintf(
			"While performing schema-based validation, an unexpected error occurred. "+
				"The attribute declares a String values validator, however its values do not implement types.StringType "+
				"or the types.StringTypable interface for custom String types. "+
				"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
				"Path: %s\nElement Type: %T", req.Path.String(), req.ConfigValue.ElementType(ctx),
		)
		tflog.Error(ctx, msg, map[string]interface{}{
			"internal_error_code": plugin.ERR_VALIDATOR_ENUM_STRINGLIST,
			"path":                req.Path.String(),
			"element_type":        fmt.Sprintf("%T", req.ConfigValue.ElementType(ctx)),
		})
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid Validator for Element Type", msg)
		return
	}

	for i, element := range req.ConfigValue.Elements() {
		elementPath := req.Path.AtListIndex(i)

		elementValuable, ok := element.(basetypes.StringValuable)
		if !ok {
			// this should *never* happen - but we want to be sure
			msg := fmt.Sprintf(
				"While performing schema-based validation, an unexpected error occurred. "+
					"The attribute declares a String values validator, however its values do not implement types.StringType "+
					"or the types.StringTypable interface for custom String types. "+
					"This is likely an issue with terraform-plugin-framework and should be reported to the provider "+
					"developers.\n\nPath: %s\nElement Type: %T\nElement Value Type: %T",
				req.Path.String(), req.ConfigValue.ElementType(ctx), element,
			)
			tflog.Error(ctx, msg, map[string]interface{}{
				"internal_error_code": plugin.ERR_VALIDATOR_ENUM_STRINGLIST,
				"path":                req.Path.String(),
				"element_type":        fmt.Sprintf("%T", req.ConfigValue.ElementType(ctx)),
				"element_value_type":  fmt.Sprintf("%T", element),
			})
			resp.Diagnostics.AddAttributeError(req.Path, "Invalid Validator for Element Value", msg)
			return
		}

		elementValue, diag := elementValuable.ToStringValue(ctx)
		resp.Diagnostics.Append(diag...)
		if resp.Diagnostics.HasError() {
			return
		}

		e := elementValue.ValueString()
		for _, val := range v.values {
			if v.ignoreCase {
				if strings.EqualFold(e, val) {
					return
				}
			} else if e == val {
				return
			}
		}
		msg := fmt.Sprintf("Value must be one of: %s", strings.Join(v.values, ", "))
		tflog.Error(ctx, fmt.Sprintf("Attribute validation failed\n\nError: %s\nAttribute: %s",
			msg, elementPath.String()), map[string]interface{}{
			"error":               msg,
			"attribute":           elementPath.String(),
			"internal_error_code": plugin.ERR_VALIDATOR_ENUM_STRINGLIST,
		})
		resp.Diagnostics.AddAttributeError(elementPath, "Invalid Value Used", msg)
	}
}
