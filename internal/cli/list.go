package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/abdellatifmht/Brik/internal/catalog"
)

func listCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available templates",
		RunE: func(cmd *cobra.Command, args []string) error {
			entries, err := catalog.Load()
			if err != nil {
				return err
			}
			for _, cat := range catalog.Categories(entries) {
				fmt.Printf("\n%s\n", cat)
				for _, e := range entries {
					if e.Manifest.Category == cat {
						fmt.Printf("  %-16s %s\n", e.ID, e.Manifest.Description)
					}
				}
			}
			fmt.Println()
			return nil
		},
	}
}
