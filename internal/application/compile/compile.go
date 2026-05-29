package compile

import (
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
	ir, err := shared.CompileSpec(info.InputFile)
	domain.CheckFatalError(err, "compiling spec")

	ir.PluginConfig = info.PluginConfig
	ir.Mode = rule.InvocationMode_MODE_COMPILE

	err = shared.RunPlugin(info.PluginName, ir)
	domain.CheckFatalError(err, "running plugin")
}
