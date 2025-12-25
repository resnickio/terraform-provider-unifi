package provider

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/resnickio/unifi-go-sdk/pkg/unifi"
)

// handleSDKError converts SDK errors to terraform diagnostics.
func handleSDKError(diags *diag.Diagnostics, err error, operation, resourceType string) {
	if err == nil {
		return
	}

	switch {
	case errors.Is(err, unifi.ErrNotFound):
		diags.AddError(
			fmt.Sprintf("%s not found", resourceType),
			fmt.Sprintf("The %s was not found. It may have been deleted outside of Terraform.", resourceType),
		)
	case errors.Is(err, unifi.ErrUnauthorized):
		diags.AddError(
			"Authentication failed",
			"The provider credentials are invalid or the session has expired. "+
				"Check your username and password configuration.",
		)
	case errors.Is(err, unifi.ErrForbidden):
		diags.AddError(
			"Access denied",
			fmt.Sprintf("You do not have permission to %s this %s. "+
				"Verify your user has the required permissions.", operation, resourceType),
		)
	case errors.Is(err, unifi.ErrConflict):
		diags.AddError(
			"Resource conflict",
			fmt.Sprintf("A %s with the same configuration already exists or conflicts with an existing resource.",
				resourceType),
		)
	case errors.Is(err, unifi.ErrBadRequest):
		var apiErr *unifi.APIError
		if errors.As(err, &apiErr) {
			diags.AddError(
				"Invalid configuration",
				fmt.Sprintf("The UniFi controller rejected the configuration: %s", apiErr.Message),
			)
		} else {
			diags.AddError(
				"Invalid configuration",
				fmt.Sprintf("The UniFi controller rejected the %s configuration. "+
					"Please verify all field values are valid.", resourceType),
			)
		}
	case errors.Is(err, unifi.ErrRateLimited):
		diags.AddError(
			"Rate limited",
			"The UniFi controller rate limited this request. Please try again later.",
		)
	case errors.Is(err, unifi.ErrServerError):
		diags.AddError(
			"Controller error",
			fmt.Sprintf("The UniFi controller encountered an internal error while trying to %s the %s. "+
				"Please try again later or check the controller logs.", operation, resourceType),
		)
	case errors.Is(err, unifi.ErrServiceUnavail):
		diags.AddError(
			"Controller unavailable",
			"The UniFi controller is currently unavailable. Please verify the controller is running and accessible.",
		)
	case errors.Is(err, unifi.ErrMethodNotAllowed):
		diags.AddError(
			"Operation not supported",
			fmt.Sprintf("The UniFi controller does not support this operation for %s. "+
				"This may be due to controller version or configuration.", resourceType),
		)
	case errors.Is(err, unifi.ErrBadGateway):
		diags.AddError(
			"Bad gateway",
			"The UniFi controller returned a bad gateway error (502). "+
				"This may indicate a proxy or network issue. Please try again later.",
		)
	case errors.Is(err, unifi.ErrGatewayTimeout):
		diags.AddError(
			"Gateway timeout",
			"The UniFi controller timed out (504). "+
				"The controller may be overloaded or unresponsive. Please try again later.",
		)
	case errors.Is(err, unifi.ErrEmptyResponse):
		var emptyErr *unifi.EmptyResponseError
		if errors.As(err, &emptyErr) {
			diags.AddError(
				"Empty response",
				fmt.Sprintf("The UniFi controller returned an empty response for %s operation on %s. "+
					"This may indicate an unexpected API change.", emptyErr.Operation, emptyErr.Resource),
			)
		} else {
			diags.AddError(
				"Empty response",
				fmt.Sprintf("The UniFi controller returned an empty response when attempting to %s the %s. "+
					"This may indicate an unexpected API change.", operation, resourceType),
			)
		}
	default:
		var apiErr *unifi.APIError
		if errors.As(err, &apiErr) {
			diags.AddError(
				fmt.Sprintf("Failed to %s %s", operation, resourceType),
				fmt.Sprintf("Status: %d, Message: %s", apiErr.StatusCode, apiErr.Message),
			)
		} else {
			diags.AddError(
				fmt.Sprintf("Failed to %s %s", operation, resourceType),
				err.Error(),
			)
		}
	}
}

// isNotFoundError checks if the error is a not found error.
func isNotFoundError(err error) bool {
	return errors.Is(err, unifi.ErrNotFound)
}

// stringPtr converts a string to a pointer, returning nil for empty strings.
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// intPtr converts a Terraform int64 to a UniFi SDK int pointer.
// Terraform uses int64 for all integers, while the UniFi SDK uses int.
// This function bridges that type difference for SDK struct population.
func intPtr(i int64) *int {
	v := int(i)
	return &v
}

// boolPtr converts a bool to a pointer.
func boolPtr(b bool) *bool {
	return &b
}

// derefString safely dereferences a string pointer, returning empty string if nil.
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// derefInt converts a UniFi SDK int pointer to a Terraform int64.
// Terraform uses int64 for all integers, while the UniFi SDK uses int.
// This function bridges that type difference for Terraform state population.
// Returns 0 if the pointer is nil.
func derefInt(i *int) int64 {
	if i == nil {
		return 0
	}
	return int64(*i)
}

// derefBool safely dereferences a bool pointer, returning false if nil.
func derefBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// stringValueOrNull returns types.StringNull() for empty strings,
// otherwise returns types.StringValue(s). Use this for optional string
// fields in SDK responses to prevent drift from empty string vs null.
func stringValueOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}
