package transfers

// TODO: Simplify the input structures
// They should really be the simplest and smallest unit that the rebalancer can reliably work with.

type Position struct {
	ID         string
	Account    Account
	Instrument Fund
	Value      Value
}

type Account struct {
	ID   string
	Name string
}

// type Instrument interface {
// 	Name() string
// 	Currency() string
// 	ISIN() string
// }

type BaseInstrument struct {
	Name     string
	Currency string
	ISIN     string
	Type     string
}

// func (i baseInstrument) Name() string {
// 	return i.name
// }

// func (i baseInstrument) Currency() string {
// 	return i.currency
// }

// func (i baseInstrument) ISIN() string {
// 	return i.isin
// }

type Fund struct {
	BaseInstrument
}

type Value struct {
	Value            float64
	Unit             string
	UnitType         string
	DecimalPrecision int
}

type Distribution struct {
	// Human-readable name of the instrument
	InstrumentName string
	// ???
	Amount int
	// Decimal percentage of the instrument, e.g. 0.15
	Distribution float64
}
