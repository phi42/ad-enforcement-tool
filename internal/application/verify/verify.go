package verify

import (
	"log/slog"

	"github.com/phi42/ad-enforcement-tool/internal/application/shared"
	"github.com/phi42/ad-enforcement-tool/internal/domain"
	"github.com/phi42/ad-enforcement-tool/rule"
)

type VerifyInfo struct {
	InputFile    string
	PluginConfig map[string]string
	PluginName   string
}

func Verify(info VerifyInfo) {
	slog.Debug("starting verify", "file", info.InputFile)

	ir, err := shared.CompileSpec(info.InputFile)
	domain.CheckFatalError(err, "loading spec")

	ir.PluginConfig = info.PluginConfig
	ir.Mode = rule.InvocationMode_MODE_VERIFY

	slog.Debug("executing plugin", "plugin", info.PluginName)

	err = shared.RunPlugin(info.PluginName, ir)
	domain.CheckFatalError(err, "running plugin")

	slog.Debug("plugin finished", "plugin", info.PluginName)
}
