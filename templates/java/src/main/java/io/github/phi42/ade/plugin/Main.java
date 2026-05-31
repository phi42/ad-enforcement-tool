package io.github.phi42.ade.plugin;

import io.github.phi42.ade.rule.Spec;

import java.io.IOException;
import java.util.Arrays;
import java.util.stream.Collectors;

/**
 * Minimal ADE enforcement plugin skeleton written in Java.
 *
 * <p>This is the minimal skeleton needed to qualify as an ADE plugin:
 * <ul>
 *   <li>Responds to {@code --info} with a JSON object listing supported modes.</li>
 *   <li>Reads a serialised {@code rule.Spec} from stdin and prints a brief summary.</li>
 * </ul>
 *
 * <p>Replace the {@link #printSpec} call with your actual compile / verify logic.
 */
public class Main {

    /** Declares which ADE invocation modes this plugin supports and the
     *  config key prefix used in .rule files for plugin-specific options. */
    record PluginInfo(String[] modes, String configPrefix) {
        String toJson() {
            String modeList = Arrays.stream(modes)
                    .map(m -> "\"" + m + "\"")
                    .collect(Collectors.joining(","));
            return "{\"modes\":[" + modeList + "],\"config_prefix\":\"" + configPrefix + "\"}";
        }
    }

    // TODO: remove any modes your plugin does not implement.
    // TODO: change "my-plugin" to the prefix your plugin reads from plugin_config.
    private static final PluginInfo INFO = new PluginInfo(new String[]{"compile", "verify"}, "my-plugin");

    public static void main(String[] args) throws IOException {
        // --info must be answered before anything else.
        // ADE calls this before every invocation to verify the plugin supports the requested mode.
        if (args.length == 1 && args[0].equals("--info")) {
            System.out.println(INFO.toJson());
            System.exit(0);
        }

        // When stdin is a terminal the user ran the binary directly; show usage.
        if (System.console() != null) {
            System.err.println("Usage: pipe an ADE Spec protobuf message to stdin");
            System.err.println("       plugin --info");
            System.exit(0);
        }

        byte[] data = System.in.readAllBytes();

        Spec spec;
        try {
            spec = Spec.parseFrom(data);
        } catch (Exception e) {
            System.err.println("error: cannot unmarshal Spec protobuf: " + e.getMessage());
            System.exit(1);
            return;
        }

        // TODO: replace this with your actual plugin logic.
        printSpec(spec);
    }

    /**
     * Prints a brief summary to confirm the plugin received and deserialised
     * the Spec correctly.  Remove this once you have real logic.
     */
    private static void printSpec(Spec spec) {
        System.out.printf("received Spec: ADR [%s] \"%s\" -- %d selector(s), %d rule(s), mode=%s%n",
                spec.getAdr().getId(),
                spec.getAdr().getTitle(),
                spec.getSelectorsList().size(),
                spec.getRulesList().size(),
                spec.getMode());
    }
}
