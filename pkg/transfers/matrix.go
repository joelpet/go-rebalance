package transfers

import (
	"errors"
	"fmt"
	"strings"
)

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
