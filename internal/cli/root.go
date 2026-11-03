// Package cli defines devkit's Cobra commands.
package cli

import "github.com/spf13/cobra"

var Version = "dev" // overridden at build time with -ldflags "-X .../cli.Version=v0.1.0"

func Execute() error {
	root := &cobra.Command{
		Use:           "devkit",
		Short:         "Generate projects, configure Git and .env, and check your tools",
		Version:       Version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(newCmd(), listCmd(), doctorCmd())
	return root.Execute()
}
