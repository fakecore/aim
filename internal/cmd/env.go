package cmd

import (
	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:   "env <model>",
	Short: "Export environment variables",
	Long:  `Export environment variables for a specific model.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement in phase 2
		return nil
	},
}

func init() {
	envCmd.Flags().String("format", "bash", "Output format: bash, fish, json")
	envCmd.Flags().String("tool", "", "Specify tool")
	envCmd.Flags().String("provider", "", "Specify provider")
}
