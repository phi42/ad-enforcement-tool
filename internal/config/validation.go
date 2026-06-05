package config

import (
	"fmt"
	"slices"
	"strings"
)

// ValidateKey rejects keys that are neither in KnownDefaults nor under the
// plugin_configs namespace.
func ValidateKey(key string) error {
	if slices.Contains(KnownDefaults, key) {
		return nil
	}
	if strings.HasPrefix(key, PluginConfigsPrefix) {
		return nil
	}
	var b strings.Builder
	fmt.Fprintf(&b, "unknown config key %q\nAllowed keys:\n", key)
	for _, k := range KnownDefaults {
		fmt.Fprintf(&b, "  %s\n", k)
	}
	fmt.Fprintf(&b, "  %s<prefix>.<key>  (plugin-specific config)", PluginConfigsPrefix)
	return fmt.Errorf("%s", b.String())
}
