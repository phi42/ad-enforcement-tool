using System.Text.Json;
using System.Text.Json.Serialization;
using Ade.Rule;
using Google.Protobuf;

// PluginInfo declares which ADE invocation modes this plugin supports and the
// config key prefix used in .rule files for plugin-specific options.
// TODO: remove any modes your plugin does not implement.
// TODO: set ConfigPrefix to the prefix your plugin reads from plugin_config.
var info = new PluginInfo { Modes = ["compile", "verify"], ConfigPrefix = "my-plugin" };

// --info must be answered before anything else.
// ADE calls this before every invocation to verify the plugin supports the requested mode.
if (args.Length == 1 && args[0] == "--info")
{
    Console.WriteLine(JsonSerializer.Serialize(info));
    return 0;
}

// When stdin is a terminal the user ran the binary directly; show usage.
if (!Console.IsInputRedirected)
{
    Console.Error.WriteLine("Usage: pipe an ADE Spec protobuf message to stdin");
    Console.Error.WriteLine("       plugin --info");
    return 0;
}

// Read the raw protobuf bytes from stdin.
byte[] data;
using (var ms = new MemoryStream())
{
    await Console.OpenStandardInput().CopyToAsync(ms);
    data = ms.ToArray();
}

Spec spec;
try
{
    spec = Spec.Parser.ParseFrom(data);
}
catch (InvalidProtocolBufferException ex)
{
    Console.Error.WriteLine($"error: cannot unmarshal Spec protobuf: {ex.Message}");
    return 1;
}

// TODO: replace this with your actual plugin logic.
PrintSpec(spec);
return 0;

// PrintSpec prints a brief summary to confirm the plugin received and
// deserialised the Spec correctly.  Remove this once you have real logic.
static void PrintSpec(Spec spec)
{
    Console.WriteLine(
        $"received Spec: ADR [{spec.Adr?.Id}] \"{spec.Adr?.Title}\" -- " +
        $"{spec.Selectors.Count} selector(s), {spec.Rules.Count} rule(s), mode={spec.Mode}");
}

internal sealed class PluginInfo
{
    [JsonPropertyName("modes")]
    public required string[] Modes { get; init; }

    [JsonPropertyName("config_prefix")]
    public required string ConfigPrefix { get; init; }
}
