package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.joelpet.se/joelpet/go-rebalance/internal/buildinfo"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version infomation.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(buildinfo.Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
