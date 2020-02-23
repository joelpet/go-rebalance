package cli

import (
	"log"

	"github.com/spf13/cobra"
	"gitlab.joelpet.se/joelpet/go-rebalance/pkg/avanza"
	"gitlab.joelpet.se/joelpet/go-rebalance/pkg/transfers"
)

var avanzaCalculateCmd = &cobra.Command{
	Use:   "calculate",
	Short: "Calculate transfers for rebalancing positions on an Avanza account according to its monthly savings distribution.",
	Run: func(cmd *cobra.Command, args []string) {
		instrumentPositionsFile, err := avanzaInstrumentPositionsCacheFile(username)
		if err != nil {
			log.Fatal(err)
		}
		positions, err := avanza.ReadAllPositions(instrumentPositionsFile)
		if err != nil {
			log.Fatal(err)
		}
		positions = avanza.FilterPositions(positions, accountID)

		monthlySavingsFile, err := avanzaMonthlySavingsCacheFile(username)
		if err != nil {
			log.Fatal(err)
		}
		distribution, err := avanza.ReadDistribution(monthlySavingsFile, accountID)
		if err != nil {
			log.Fatal(err)
		}

		transfers.Calculate(positions, distribution)
	},
}

var (
	accountID string
)

func init() {
	avanzaCmd.AddCommand(avanzaCalculateCmd)

	avanzaCalculateCmd.
		Flags().
		StringVar(&accountID, "account-id", "", "id of the account to calculate rebalancing transfers for")

	avanzaCalculateCmd.MarkFlagRequired("account-id")
}
