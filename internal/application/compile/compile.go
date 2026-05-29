package compile

import (
	"log/slog"

	"github.com/phi42/ad-enforcement-tool/internal/application/shared"
	"github.com/phi42/ad-enforcement-tool/internal/domain"
	"github.com/phi42/ad-enforcement-tool/rule"
)

type CompileInfo struct {
	InputFile    string
	PluginConfig map[string]string
	PluginName   string
}

func Compile(info CompileInfo) {
	slog.Debug("starting compilation", "file", info.InputFile)

	ir, err := shared.CompileSpec(info.InputFile)
	domain.CheckFatalError(err, "compiling spec")

	ir.PluginConfig = info.PluginConfig
	ir.Mode = rule.InvocationMode_MODE_COMPILE

	slog.Debug("executing plugin", "plugin", info.PluginName)

	err = shared.RunPlugin(info.PluginName, ir)
	domain.CheckFatalError(err, "running plugin")

	slog.Debug("plugin finished", "plugin", info.PluginName)
}
