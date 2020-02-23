package cli

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/spf13/cobra"
	"gitlab.joelpet.se/joelpet/go-rebalance/pkg/avanza"
)

var avanzaFetchCmd = &cobra.Command{
	Use: "fetch",
	Run: func(cmd *cobra.Command, args []string) {
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
		if monthlySavings, err := azaclt.GetMonthlySavings(); err != nil {
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

var (
	password string
	totp     string
)

func init() {
	avanzaCmd.AddCommand(avanzaFetchCmd)

	// TODO: Read password and TOTP securely (environment or prompt)
	avanzaFetchCmd.
		Flags().
		StringVar(&password, "password", "", "Password for authenticating")
	avanzaFetchCmd.
		Flags().
		StringVar(&totp, "totp", "", "TOTP for authenticating")

	avanzaFetchCmd.MarkFlagRequired("password")
	avanzaFetchCmd.MarkFlagRequired("totp")
}
