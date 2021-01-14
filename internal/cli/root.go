package cli

import (
	"log"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "rebalance",
}

func Execute() {
	rootCmd.Execute()
}

func init() {
	log.SetFlags(log.LstdFlags | log.Llongfile)

}
