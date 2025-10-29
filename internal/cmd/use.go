package cmd

import (
	"github.com/spf13/cobra"
)

var useCmd = &cobra.Command{
	Use:   "use <model>",
	Short: "Switch to a model",
	Long:  `Switch to a specific AI model and provider.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement in phase 2
		return nil
	},
}

func init() {
	useCmd.Flags().Bool("global", false, "Set global default")
	useCmd.Flags().Bool("local", false, "Set project default")
	useCmd.Flags().String("tool", "", "Specify tool")
	useCmd.Flags().String("provider", "", "Specify provider")
	useCmd.Flags().Bool("export", false, "Export environment variables")
}
