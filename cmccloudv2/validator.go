package cmccloudv2

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// Hàm để tùy chỉnh diff và bỏ qua các thay đổi của các thuộc tính cụ thể
func ignoreChangesCustomizeDiff(fieldsToIgnore ...string) schema.CustomizeDiffFunc {
	return func(diff *schema.ResourceDiff, v interface{}) error {
		for _, field := range fieldsToIgnore {
			if diff.HasChange(field) {
				diff.Clear(field)
			}
		}
		return nil
	}
}

func validateBillingMode(v interface{}, key string) (warnings []string, errors []error) {
	biling_mode := v.(string)
	if biling_mode != "monthly" && biling_mode != "hourly" {
		return nil, []error{fmt.Errorf("%s must be one of two values: `monthly` or `hourly`", key)}
	}
	return nil, nil
}

func validateNetmask(v interface{}, k string) (warnings []string, errors []error) {
	re := `^(((255\.){3}(255|254|252|248|240|224|192|128|0+))|((255\.){2}(255|254|252|248|240|224|192|128|0+)\.0)|((255\.)(255|254|252|248|240|224|192|128|0+)(\.0+){2})|((255|254|252|248|240|224|192|128|0+)(\.0+){3}))$`
	return validateRegexp(re)(v, k)
}
func validateUUID(v interface{}, k string) (warnings []string, errors []error) {
	re := `^[0-9a-fA-F]{8}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{12}$`
	return validateRegexp(re)(v, k)
}

func validateFirewallID(v interface{}, k string) (warnings []string, errors []error) {
	re := `^(allow|deny|[0-9a-fA-F]{8}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{12})$`
	return validateRegexp(re)(v, k)
}
func validatePortNumber(val interface{}, key string) (warns []string, errs []error) {
	v, ok := val.(int)
	if !ok {
		errs = append(errs, fmt.Errorf("%q must be an integer", key))
		return
	}

	if v < 1 || v > 65535 {
		errs = append(errs, fmt.Errorf("%q must be between 1 and 65535, got %d", key, v))
	}
	return
}

// validatePortRange checks if the string is a valid port range in the format portstart-portend
func validatePortRange(val interface{}, key string) (warns []string, errs []error) {
	v, ok := val.(string)
	if !ok {
		errs = append(errs, fmt.Errorf("%q is not a valid string", key))
		return
	}

	// Split the string by hyphen
	parts := strings.Split(v, "-")
	if len(parts) != 2 {
		errs = append(errs, fmt.Errorf("%q must be in the format portstart-portend", key))
		return
	}

	// Parse the start and end ports
	startPort, err := strconv.Atoi(parts[0])
	if err != nil {
		errs = append(errs, fmt.Errorf("%q start port is not a valid number", key))
		return
	}

	endPort, err := strconv.Atoi(parts[1])
	if err != nil {
		errs = append(errs, fmt.Errorf("%q end port is not a valid number", key))
		return
	}

	// Validate the port numbers
	if startPort < 1 || startPort > 65535 {
		errs = append(errs, fmt.Errorf("%q start port must be between 1 and 65535", key))
	}

	if endPort < 1 || endPort > 65535 {
		errs = append(errs, fmt.Errorf("%q end port must be between 1 and 65535", key))
	}

	// Validate the port range
	if startPort > endPort {
		errs = append(errs, fmt.Errorf("%q start port must be less than or equal to end port", key))
	}

	return
}
func validateRegexp(re string) schema.SchemaValidateFunc {
	return func(v interface{}, k string) (ws []string, errors []error) {
		value := fmt.Sprint(v)
		if !regexp.MustCompile(re).MatchString(value) {
			errors = append(errors, fmt.Errorf(
				"%q (%q) doesn't match regexp %q", k, value, re))
		}
		return
	}
}

func validateIPCidrRange(v interface{}, k string) (warnings []string, errors []error) {
	_, _, err := net.ParseCIDR(v.(string))
	if err != nil {
		errors = append(errors, fmt.Errorf("%q is not a valid IP CIDR range: %s", k, err))
	}
	return
}
func validateIPAddress(i interface{}, val string) ([]string, []error) {
	ip := net.ParseIP(i.(string))
	if ip == nil {
		return nil, []error{fmt.Errorf("could not parse %q to IP address", val)}
	}
	return nil, nil
}
func validateAll(validators ...schema.SchemaValidateFunc) schema.SchemaValidateFunc {
	return func(val interface{}, key string) (warns []string, errs []error) {
		for _, validator := range validators {
			w, e := validator(val, key)
			warns = append(warns, w...)
			errs = append(errs, e...)
		}
		return warns, errs
	}
}
func validateEmpty(val interface{}, key string) (warns []string, errs []error) {
	v, ok := val.(string)
	if !ok || v != "" {
		errs = append(errs, fmt.Errorf("%q must be an empty string", key))
	}
	return
}
func validateAny(error_message string, validators ...schema.SchemaValidateFunc) schema.SchemaValidateFunc {
	return func(val interface{}, key string) (warns []string, errs []error) {
		for _, validator := range validators {
			_, e := validator(val, key)
			if len(e) == 0 {
				// If any validation function returns no errors, validation passes
				return nil, nil
			}
		}
		// If all validation functions fail, return the errors from the last validation
		errs = append(errs, fmt.Errorf("%q "+error_message, key))
		return nil, errs
	}
}
