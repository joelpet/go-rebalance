package avanza

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"

	"gitlab.joelpet.se/joelpet/go-rebalance/pkg/transfers"
)

// ReadPositions reads positions from file including all accounts.
func ReadAllPositions(filename string) ([]transfers.Position, error) {
	var azapos PositionsPayload
	if contents, err := ioutil.ReadFile(filename); err != nil {
		return nil, fmt.Errorf("avanza: reading positions file: %s", err)
	} else if err := json.Unmarshal(contents, &azapos); err != nil {
		return nil, fmt.Errorf("avanza: unmarshalling positions: %s", err)
	}

	var positions []transfers.Position
	for _, p := range azapos.InstrumentPositions {
		for _, p := range p.Positions {
			position := transfers.Position{
				Account: transfers.Account{
					ID:   p.AccountID,
					Name: p.AccountName,
				},
				Instrument: transfers.Fund{
					BaseInstrument: transfers.BaseInstrument{
						Name:     p.Name,
						Currency: p.Currency,
					},
				},
				Value: transfers.Value{
					Value: float64(p.Value),
					Unit:  p.Currency,
				},
			}
			positions = append(positions, position)
		}
	}

	return positions, nil
}

// FilterPositions returns a slice with only those positions matching a given account id.
func FilterPositions(positions []transfers.Position, accountID string) []transfers.Position {
	filtered := make([]transfers.Position, 0, len(positions))
	for _, p := range positions {
		if p.Account.ID == accountID {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

func ReadDistribution(filename string, accountID string) ([]transfers.Distribution, error) {
	var azadist PeriodicSavingsPayload
	if contents, err := ioutil.ReadFile(filename); err != nil {
		return nil, fmt.Errorf("avanza: reading monthly savings file: %s", err)
	} else if err := json.Unmarshal(contents, &azadist); err != nil {
		return nil, fmt.Errorf("avanza: unmarshalling monthly savings: %s", err)
	}

	var distributions []transfers.Distribution
	for _, ps := range azadist.PeriodicSavings {
		if strconv.Itoa(ps.Account.AccountID) == accountID {
			for _, av := range ps.AllocationViews {
				distribution := transfers.Distribution{
					InstrumentName: av.Name,
					Distribution:   float64(av.Allocation) / 100,
				}
				distributions = append(distributions, distribution)
			}
		}
	}

	return distributions, nil
}
