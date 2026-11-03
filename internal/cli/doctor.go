package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/abdellatifmht/Brik/internal/doctor"
)

func doctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check that development tools are installed",
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, t := range doctor.Tools {
				if v, ok := doctor.Version(t); ok {
					fmt.Printf("✔ %-8s %s\n", t.Cmd, v)
				} else {
					fmt.Printf("✘ %-8s absent → %s\n", t.Cmd, doctor.Hint(t.Cmd))
				}
			}
			return nil
		},
	}
}
