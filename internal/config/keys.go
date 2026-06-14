package config

// File name and extension of the project-level config file.
//
// The full filename on disk is FileName + "." + FileExt, i.e. ".ade.yaml".
const (
	FileName = ".ade"
	FileExt  = "yaml"
)

// Top-level YAML key prefixes.
const (
	// PluginLocationsPrefix is the dotted path prefix for installed plugin
	// binary paths: plugin_locations.<name> -> /path/to/plugin.
	PluginLocationsPrefix = "plugin_locations."

	// PluginConfigsPrefix is the dotted path prefix for plugin-specific
	// configuration: plugin_configs.<config_prefix>.<key> -> value.
	// The middle segment must match a plugin's advertised config_prefix.
	PluginConfigsPrefix = "plugin_configs."

	// PluginSourcesPrefix is the dotted path prefix used to remember the
	// remote module URL a plugin was installed from, so that
	// `ade plugin update` can re-fetch it.
	PluginSourcesPrefix = "plugin_sources."
)

// Default-value keys that can be set with `ade config set`.
const (
	DefaultInput = "defaults.input"

	DefaultCompilePlugin = "defaults.compile.plugin"
	DefaultCompileInput  = "defaults.compile.input"
	DefaultVerifyPlugin  = "defaults.verify.plugin"
	DefaultVerifyInput   = "defaults.verify.input"
)

// KnownDefaults is the closed list of dotted keys recognised by `ade config`
// for default values. Plugin-specific keys (PluginConfigsPrefix + ...) are
// allowed in addition to these but are open-ended and not enumerated here.
var KnownDefaults = []string{
	DefaultInput,
	DefaultCompilePlugin,
	DefaultCompileInput,
	DefaultVerifyPlugin,
	DefaultVerifyInput,
}
