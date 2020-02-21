package main

import (
	"errors"
	"fmt"
	"github.com/lanl/clp"
	"math"
	"strings"
)

var ()

// A:	+4099	Spiltan Aktiefond Investmentbolag
// B:	+603	DNB Global Indeks A
// C:	+10095	SPP Aktiefond USA
// D:	-345	Avanza Zero
// E:	-3982	Länsförsäkringar Tillväxtmrkd Idxnära A
// F:	+1019	Swedbank Robur Access USA
// G:	+868	Länsförsäkringar Global Indexnära
// H:	-1845	Öhman Etisk Emerging Markets A
// I:	-802	Länsförsäkringar Europa Indexnära
// J:	-798	SPP Aktiefond Europa
// K:	-8913	Avanza USA

func main() {
	instruments := []string{
		"Spiltan Aktiefond Investmentbolag",
		"DNB Global Indeks A",
		"SPP Aktiefond USA",
		"Avanza Zero",
		"Länsförsäkringar Tillväxtmrkd Idxnära A",
		"Swedbank Robur Access USA",
		"Länsförsäkringar Global Indexnära",
		"Öhman Etisk Emerging Markets A",
		"Länsförsäkringar Europa Indexnära",
		"SPP Aktiefond Europa",
		"Avanza USA",
	}
	// transfers are the vars defining a specific instance of the optimization problem
	transfers := make([]string, 0, len(instruments)*(len(instruments)-1))
	for i, from := range instruments {
		for j, to := range instruments {
			if i != j {
				transfers = append(transfers, fmt.Sprintf("%-33s -> %-39s", from, to))
			}
		}
	}

	// Set up the optimization problem.
	pinf := math.Inf(1)
	// ninf := math.Inf(-1)
	simp := clp.NewSimplex()

	balances := []int{
		4099,
		603,
		10095,
		-345,
		-3982,
		1019,
		868,
		-1845,
		-802,
		-798,
		-8913,
	}

	nVars := len(balances) * (len(balances) - 1)

	obj := make([]float64, nVars)
	for i := range obj {
		obj[i] = 1.0
	}

	varBounds := make([][2]float64, nVars)
	for i := range varBounds {
		varBounds[i] = [2]float64{0, pinf}
	}

	// ineqs := make([][]float64, len(balances))
	ineqs := NewMatrix(len(balances), nVars+2) // ||{lb,up}|| == 2
	block := NewMatrix(len(balances), len(balances)-1)
	ones := NewMatrix(1, block.columnsCount()).withValues(1.0)
	eye := NewMatrix(len(balances)-1, len(balances)-1).withEyeValues(-1.0)
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
	for i, balance := range balances {
		bounds.set(i, 0, float64(balance))
	}
	ineqs.setBlock(bounds, 0, 0)
	ineqs.setBlock(bounds, 0, ineqs.columnsCount()-1)

	simp.EasyLoadDenseProblem(
		//        A>B  A>C  A>D  A>E  A>F  A>G  A>H  A>I  A>J  A>K  A>L, B>A, B>C, B>D, ..., K>A, K>B, K>C, K>D, K>E, K>F, K>G, K>H, K>I, K>J
		// []float64{1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0},
		obj,
		// [][2]float64{
		// 	// LB UB
		// 	{0, pinf}, // 1 ≤ a ≤ ∞
		// 	{0, pinf}, // 1 ≤ b ≤ ∞
		// 	{0, pinf}, // 1 ≤ c ≤ ∞
		// },
		varBounds,
		ineqs.nestedSlices(),
		// [][]float64{
		// 	// LB  a    b    c  ... x   y   z   (... x_110)    UB
		// 	// LB  a    b    c    UB
		// 	{1.0, 1.0, -1.0, 0.0, pinf},  // 1 ≤ a - b ≤ ∞
		// 	{1.0, 0.0, 1.0, -1.0, pinf},  // 1 ≤ b - c ≤ ∞
		// 	{ninf, 1.0, -2.0, 1.0, -1.0}, // -∞ ≤ a - 2b + c ≤ -1
		// }
	)

	simp.SetOptimizationDirection(clp.Minimize)

	// Solve the optimization problem.
	simp.Primal(clp.NoValuesPass, clp.NoStartFinishOptions)
	soln := simp.PrimalColumnSolution()

	// Output the solution.
	for i, amount := range soln {
		if amount != 0 {
			fmt.Printf("%s : %9.2f\n", transfers[i], amount)
		}
	}
}

type matrix struct {
	elems [][]float64
}

// NewMatrix creates a new zeroed m✕n matrix, i.e. one with m rows and n columns of zeroes.
func NewMatrix(m, n int) matrix {
	matrix := matrix{}
	matrix.elems = make([][]float64, m)
	for i := range matrix.elems {
		matrix.elems[i] = make([]float64, n)
	}
	return matrix
}

func (m matrix) withValues(value float64) matrix {
	for i := range m.elems {
		for j := range m.elems[i] {
			m.elems[i][j] = value
		}
	}
	return m
}

func (m matrix) withEyeValues(value float64) matrix {
	for i := 0; i < m.rowsCount() && i < m.columnsCount(); i++ {
		m.elems[i][i] = value
	}
	return m
}

func (m matrix) rowsCount() int {
	return len(m.elems)
}

func (m matrix) columnsCount() int {
	return len(m.elems[0])
}

func (m matrix) nestedSlices() [][]float64 {
	return m.elems
}

// swapRows swaps the contents on row i with row j
func (m matrix) swapRows(i, j int) error {
	if i < 0 || i >= len(m.elems) || j < 0 || j >= len(m.elems) {
		return errors.New("invalid index")
	}
	tmp := m.elems[i]
	m.elems[i] = m.elems[j]
	m.elems[j] = tmp
	return nil
}

func (m matrix) set(i, j int, value float64) error {
	if i < 0 || i >= m.rowsCount() || j < 0 || j >= m.columnsCount() {
		return errors.New("index out of bounds")
	}
	m.elems[i][j] = value
	return nil
}

// setBlock sets the values of a submatrix in m to the values from o. The
// submatrix is located with its top left corner at (rowOffset, colOffset) and
// has dimensions o.rowsCount()✕o.columnsCount()
func (m matrix) setBlock(o matrix, rowOffset, colOffset int) error {
	if colOffset < 0 || rowOffset < 0 {
		return errors.New("start position outside target matrix")
	}
	if lastColumnIdx := o.columnsCount() + colOffset - 1; lastColumnIdx >= m.columnsCount() {
		return errors.New("other matrix overflows horizontally")
	}
	if lastRowIdx := o.rowsCount() + rowOffset - 1; lastRowIdx >= m.rowsCount() {
		return errors.New("other matrix overflows vertically")
	}
	for i := 0; i < o.rowsCount(); i++ {
		for j := 0; j < o.columnsCount(); j++ {
			m.elems[i+rowOffset][j+colOffset] = o.elems[i][j]
		}
	}
	return nil
}

func (m matrix) String() string {
	var str strings.Builder
	for i := 0; i < m.rowsCount(); i++ {
		for j := 0; j < m.columnsCount(); j++ {
			str.WriteString(fmt.Sprintf("%2.f ", m.elems[i][j]))
		}
		str.WriteString("\n")
	}
	return str.String()
}
