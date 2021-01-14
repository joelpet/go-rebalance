package transfers

import (
	"errors"
	"fmt"
	"log"
	"math"

	"github.com/lanl/clp"
)

// Calculate finds a smallest set of amounts to transfer that balances the given
// deviations and outputs the result to stdout.
func Calculate(positions []Position, distributions []Distribution) {
	switch {
	case len(positions) == 0:
		log.Fatal("No positions to rebalance")
	case len(distributions) == 0:
		log.Fatal("Empty target distribution")
	}

	balances, err := calculateBalances(positions, distributions)
	if err != nil {
		log.Fatal(err)
	}
	balancer := newBalancer(balances)
	transfers := balancer.optimalTransfers()

	fmt.Printf("# Current positions (# %d)\n", len(positions))
	for _, p := range positions {
		fmt.Printf("%-45s: %10.2f %s\n", p.Instrument.Name, p.Value.Value, p.Value.Unit)
	}
	fmt.Println()

	dSum := 0.0
	for _, d := range distributions {
		dSum += 100 * d.Distribution
	}
	fmt.Printf("# Target distribution (%f %%)\n", dSum)
	for _, d := range distributions {
		fmt.Printf("%-45s: %6.2f %%\n", d.InstrumentName, 100*d.Distribution)
	}
	fmt.Println()

	bSum := 0.0
	for _, balance := range balances {
		bSum += balance
	}
	fmt.Printf("# Calculated deviations (∑ %f)\n", bSum)
	for instr, balance := range balances {
		fmt.Printf("%-45s: %10.2f\n", instr, balance)
	}
	fmt.Println()

	positionValue := map[string]Value{}
	for _, pos := range positions {
		positionValue[pos.Instrument.Name] = pos.Value
	}

	fmt.Printf("# Calculated transfers (# %d)\n", len(transfers))
	for _, t := range transfers {
		volume := t.amount / positionValue[t.from].Value * 100
		fmt.Printf("%-45s -> %-45s : %10.2f   (%20.16f %%)\n", t.from, t.to, t.amount, volume)
	}
}

type positionVerifier struct {
	instrumentCurrency string
	valueUnit          string
	valueUnitType      string

	errs []error
}

// newPositionVerifierSample creates a new positionVerifier based on an
// arbitrarily selected sample from the given positions slice.
func newPositionVerifierSample(positions []Position) (*positionVerifier, error) {
	if len(positions) <= 0 {
		return nil, errors.New("no positions to sample from")
	}
	p := positions[0]
	return &positionVerifier{
		instrumentCurrency: p.Instrument.Currency,
		valueUnit:          p.Value.Unit,
		valueUnitType:      p.Value.UnitType,
	}, nil
}

func (v positionVerifier) inspect(p Position) {
	if v.instrumentCurrency != p.Instrument.Currency {
		v.errs = append(v.errs, ErrCurrency)
	}
	if v.valueUnit != p.Value.Unit {
		v.errs = append(v.errs, ErrUnit)
	}
	if v.valueUnitType != p.Value.UnitType {
		v.errs = append(v.errs, ErrUnitType)
	}
}

func (v positionVerifier) errors() []error {
	return v.errs
}

// Errors that may occur during position verification.
var (
	ErrCurrency = errors.New("currency differs")
	ErrUnit     = errors.New("unit differs")
	ErrUnitType = errors.New("unit type differs")
)

type positionSummer struct {
	sum float64
}

func (s *positionSummer) add(p Position) {
	s.sum += p.Value.Value
}

func (s positionSummer) total() float64 {
	return s.sum
}

type balanceCalculator struct {
	total     float64
	instrDist map[string]Distribution

	balances map[string]float64
}

func newBalanceCalculator(totalValue float64, dists []Distribution) balanceCalculator {
	calculator := balanceCalculator{
		total:     float64(totalValue),
		instrDist: map[string]Distribution{},
		balances:  map[string]float64{},
	}
	for _, dist := range dists {
		calculator.instrDist[dist.InstrumentName] = dist
	}
	return calculator
}

func (c balanceCalculator) includePosition(pos Position) {
	instrName := pos.Instrument.Name
	targetDist := c.instrDist[instrName].Distribution
	targetValue := targetDist * c.total
	currentValue := pos.Value.Value
	c.balances[instrName] = currentValue - targetValue
}

func (c balanceCalculator) includeDistribution(dist Distribution) {
	if _, ok := c.balances[dist.InstrumentName]; !ok {
		c.includePosition(Position{
			Instrument: Fund{
				BaseInstrument: BaseInstrument{
					Name: dist.InstrumentName,
				},
			},
			Value: Value{
				Value: 0,
			},
		})
	}
}

func calculateBalances(positions []Position, distributions []Distribution) (map[string]float64, error) {
	posVerifier, err := newPositionVerifierSample(positions)
	if err != nil {
		return nil, err
	}

	posSummer := positionSummer{}
	for _, position := range positions {
		posVerifier.inspect(position)
		posSummer.add(position)
	}

	if posErrors := posVerifier.errors(); len(posErrors) > 0 {
		return nil, errors.New("irrecoverable inconsistency detected among positions")
	}

	balanceCalculator := newBalanceCalculator(posSummer.total(), distributions)
	for _, position := range positions {
		balanceCalculator.includePosition(position)
	}

	for _, distribution := range distributions {
		balanceCalculator.includeDistribution(distribution)
	}

	return balanceCalculator.balances, nil
}

type transfer struct {
	from   string
	to     string
	amount float64
}

type balancer struct {
	simplex     *clp.Simplex
	instruments []string
	deviations  []float64
}

func newBalancer(instrDevs map[string]float64) balancer {
	balancer := balancer{
		simplex:     clp.NewSimplex(),
		instruments: make([]string, 0, len(instrDevs)),
		deviations:  make([]float64, 0, len(instrDevs)),
	}
	for instr, dev := range instrDevs {
		balancer.instruments = append(balancer.instruments, instr)
		balancer.deviations = append(balancer.deviations, dev)
	}
	return balancer
}

func (b balancer) optimalTransfers() []transfer {
	b.simplex.EasyLoadDenseProblem(b.obj(), b.varBounds(), b.ineqs())
	b.simplex.SetOptimizationDirection(clp.Minimize)
	b.simplex.Primal(clp.NoValuesPass, clp.NoStartFinishOptions)
	soln := b.simplex.PrimalColumnSolution()
	return b.translateSolution(soln)
}

// translateSolution converts the given solution into a slice of transfers.
func (b balancer) translateSolution(soln []float64) []transfer {
	nonzeroCount := 0
	for _, amount := range soln {
		if amount != 0 {
			nonzeroCount++
		}
	}

	solnIdx, transfers := 0, make([]transfer, 0, nonzeroCount)
	for i, from := range b.instruments {
		for j, to := range b.instruments {
			if i != j {
				if amount := soln[solnIdx]; amount != 0 {
					t := transfer{from: from, to: to, amount: soln[solnIdx]}
					transfers = append(transfers, t)
				}
				solnIdx++
			}
		}
	}

	return transfers
}

func (b balancer) nVars() int {
	return len(b.deviations) * (len(b.deviations) - 1)
}

// obj returns the coefficients of the objective function.
func (b balancer) obj() []float64 {
	obj := make([]float64, b.nVars())
	for i := range obj {
		obj[i] = 1.0
	}
	return obj
}

// varBounds returns the lower and upper bounds on each variable.
func (b balancer) varBounds() [][2]float64 {
	varBounds := make([][2]float64, b.nVars())
	for i := range varBounds {
		varBounds[i] = [2]float64{0, math.Inf(1)}
	}
	return varBounds
}

// ineqs returns a matrix in which each row is of the form {lower bound, var_1,
// var_2, …, var_N, upper bound}.
func (b balancer) ineqs() [][]float64 {
	ineqs := NewMatrix(len(b.deviations), b.nVars()+2) // ||{lb,up}|| == 2
	block := NewMatrix(len(b.deviations), len(b.deviations)-1)
	ones := NewMatrix(1, block.columnsCount()).withValues(1.0)
	eye := NewMatrix(len(b.deviations)-1, len(b.deviations)-1).withEyeValues(-1.0)
	block.setBlock(ones, 0, 0)
	block.setBlock(eye, 1, 0)

	colOffset := 1
	for i := 0; i < block.rowsCount()-1; i++ {
		ineqs.setBlock(block, 0, colOffset)
		block.swapRows(i, i+1)
		colOffset += block.columnsCount()
	}
	ineqs.setBlock(block, 0, colOffset)

	bounds := NewMatrix(ineqs.rowsCount(), 1)
	for i, deviation := range b.deviations {
		bounds.set(i, 0, deviation)
	}
	ineqs.setBlock(bounds, 0, 0)
	ineqs.setBlock(bounds, 0, ineqs.columnsCount()-1)

	return ineqs.nestedSlices()
}
