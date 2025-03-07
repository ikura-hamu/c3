package c3

import (
	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
)

func init() {
	register.Plugin("c3", New)
}

func New(settings any) (register.LinterPlugin, error) {
	// The configuration type will be map[string]any or []interface, it depends on your configuration.
	// You can use https://github.com/go-viper/mapstructure to convert map to struct.

	return &plugin{}, nil
}

type plugin struct{}

var _ register.LinterPlugin = new(plugin)

func (*plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		Analyzer,
	}, nil
}

func (*plugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}
