package shared

import (
	"github.com/gruntwork-io/terragrunt/internal/cli/flags"
	"github.com/gruntwork-io/terragrunt/internal/clihelper"
	"github.com/gruntwork-io/terragrunt/pkg/options"
)

const (
	ProgressFlagName = "progress"
)

// NewProgressFlag creates the --progress flag for enabling the BuildKit-style progress TUI.
func NewProgressFlag(opts *options.TerragruntOptions) *flags.Flag {
	tgPrefix := flags.Prefix{flags.TgPrefix}

	return flags.NewFlag(&clihelper.BoolFlag{
		Name:        ProgressFlagName,
		EnvVars:     tgPrefix.EnvVars(ProgressFlagName),
		Destination: &opts.Progress,
		Usage:       "Show a live progress display for parallel unit execution (requires interactive terminal).",
	})
}
