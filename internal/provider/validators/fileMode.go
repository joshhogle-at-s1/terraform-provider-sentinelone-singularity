package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/joshhogle-at-s1/terraform-provider-sentinelone-singularity/internal/plugin"
)

// ensure implementation satisfied expected interfaces
var _ validator.String = fileMode{}

// FileModeIsValid returns a validator which ensurses that the value given is a valid file mode.
func FileModeIsValid() validator.String {
	return fileMode{}
}

// fileMode holds details about the file mode validator.
type fileMode struct{}

// Description returns a plain text description of the validator's behavior, suitable for a practitioner to
// understand its impact.
func (v fileMode) Description(ctx context.Context) string {
	return "checks that the value given is a valid file mode"
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior, suitable for a
// practitioner to understand its impact.
func (v fileMode) MarkdownDescription(ctx context.Context) string {
	return "checks that the value given is a valid file mode"
}

// Validate runs the main validation logic of the validator, reading configuration data out of `req` and
// updating `resp` with diagnostics.
func (v fileMode) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	_, diags := plugin.ParseFilesystemMode(ctx, req.ConfigValue.ValueString())
	if diags.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Attribute validation failed\n\nError: %s\nAttribute: %s",
			diags[0].Detail(), req.Path.String()), map[string]interface{}{
			"error":               diags[0].Detail,
			"attribute":           req.Path.String(),
			"internal_error_code": plugin.ERR_VALIDATOR_ENUM_STRING,
		})
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid Value Used", diags[0].Detail())
		return
	}
}
