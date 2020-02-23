package cli

import (
	"github.com/spf13/cobra"
	"log"
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
