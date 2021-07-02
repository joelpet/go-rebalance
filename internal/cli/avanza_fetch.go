package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/spf13/cobra"
	"gitlab.joelpet.se/joelpet/go-rebalance/pkg/avanza"
	"golang.org/x/term"
)

var avanzaFetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch account data from Avanza through the web API.",
	Run: func(cmd *cobra.Command, args []string) {
		var password string
		if password = os.Getenv("GO_REBALANCE_AVANZA_PASSWORD"); password == "" {
			fmt.Print("Password [GO_REBALANCE_AVANZA_PASSWORD]: ")
			if input, err := term.ReadPassword(int(os.Stdin.Fd())); err != nil {
				log.Fatal(err)
			} else {
				password = string(input)
				fmt.Println()
			}
		}

		var totp string
		if totp = os.Getenv("GO_REBALANCE_AVANZA_TOTP"); totp == "" {
			fmt.Print("TOTP [GO_REBALANCE_AVANZA_TOTP]: ")
			if input, err := term.ReadPassword(int(os.Stdin.Fd())); err != nil {
				log.Fatal(err)
			} else {
				totp = string(input)
				fmt.Println()
			}
		}

		var azaclt *avanza.Client
		var err error
		if azaclt, err = avanza.NewClient(); err != nil {
			log.Fatal(err)
		} else if err := azaclt.Authenticate(avanza.UserCredentials{
			Username: username, Password: password, AuthTimeout: 60}); err != nil {
			log.Fatal(err)
		} else if err := azaclt.TOTP(avanza.TOTP{
			Method: "TOTP", TOTPCode: totp}); err != nil {
			log.Fatal(err)
		}

		// Cache monthly savings
		if monthlySavings, err := azaclt.GetPeriodicSavings(); err != nil {
			log.Fatal(err)
		} else if data, err := json.Marshal(monthlySavings); err != nil {
			log.Fatal(err)
		} else if f, err := avanzaMonthlySavingsCacheFile(username); err != nil {
			log.Fatal(err)
		} else if err := ioutil.WriteFile(f, data, os.FileMode(0600)); err != nil {
			log.Fatal(err)
		}

		// Cache instrument positions
		if positions, err := azaclt.GetPositions(); err != nil {
			log.Fatal(err)
		} else if data, err := json.Marshal(positions); err != nil {
			log.Fatal(err)
		} else if f, err := avanzaInstrumentPositionsCacheFile(username); err != nil {
			log.Fatal(err)
		} else if err := ioutil.WriteFile(f, data, os.FileMode(0600)); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	avanzaCmd.AddCommand(avanzaFetchCmd)
}
