package transfers

import (
	"testing"
)

func TestCalculateBalances(t *testing.T) {
	positions := []Position{
		{
			ID:         "A",
			Instrument: Fund{BaseInstrument{Name: "A fund"}},
			Value:      Value{Value: 100.00},
		},
		{
			ID:         "B",
			Instrument: Fund{BaseInstrument{Name: "B fund"}},
			Value:      Value{Value: 200.00},
		},
		{
			ID:         "C",
			Instrument: Fund{BaseInstrument{Name: "C fund"}},
			Value:      Value{Value: 300.00},
		},
	}
	distributions := []Distribution{
		{
			InstrumentName: "A fund",
			Distribution:   0.10,
		},
		{
			InstrumentName: "B fund",
			Distribution:   0.50,
		},
		{
			InstrumentName: "C fund",
			Distribution:   0.40,
		},
	}
	balances, err := calculateBalances(positions, distributions)
	if err != nil {
		t.Fatal(err)
	}

	if want, got := 40., balances["A fund"]; want != got {
		t.Errorf("balances[A fund] = %f, want %f", got, want)
	}
	if want, got := -100., balances["B fund"]; want != got {
		t.Errorf("balances[B fund] = %f, want %f", got, want)
	}
	if want, got := 60., balances["C fund"]; want != got {
		t.Errorf("balances[C fund] = %f, want %f", got, want)
	}
}
