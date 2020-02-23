package cli

import (
	"github.com/adrg/xdg"
	"path/filepath"

	"github.com/spf13/cobra"
)

var avanzaCmd = &cobra.Command{
	Use: "avanza",
}

func avanzaInstrumentPositionsCacheFile(username string) (string, error) {
	relPath := filepath.Join("go-rebalance", "avanza", username, "instrument_positions.json")
	return xdg.CacheFile(relPath)
}

func avanzaMonthlySavingsCacheFile(username string) (string, error) {
	relPath := filepath.Join("go-rebalance", "avanza", username, "monthly_savings.json")
	return xdg.CacheFile(relPath)
}

var (
	username string
)

func init() {
	rootCmd.AddCommand(avanzaCmd)

	avanzaCmd.
		PersistentFlags().
		StringVar(&username, "username", "", "Username for authenticating")

	avanzaCmd.MarkPersistentFlagRequired("username")
}
