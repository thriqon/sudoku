// Package sudoku provides a sudoku simple and fast solver based on "Solving
// Every Sudoku Puzzle" by Peter Norvig (http://norvig.com/sudoku.html).
//
// The approach by Norvig uses constraint propagation/satisfaction and search.
// Anytime a number is assigned to a square, this number is removed from any
// 'peers' (see below for explanation).  If any of the peer then becomes filled
// by having only possibility remaining, this is propagated as well
// (recursively). This simple method of constraint propagation is enough for simple
// Sudokus, but harder ones need more work.
//
// Instead of implementing more 'intelligent' approaches, Norvig decided to
// simply try every possibility and reject conflicting moves. For optimization,
// the square with the least possibilities is filled first.
//
// Anytime a function returns a sudoku and/or an error, the sudoku is only
// valid if the error is nil.
package sudoku

import (
	"fmt"
	"io"
	"strings"
)

var (
	// ErrConflict is returned when there is a conflict that prevents finding a
	// solution or assigning a value.
	ErrConflict = fmt.Errorf("Conflict")
)

// A Sudoku is an immutable value, it contains the 81 fields of a standard
// playing field. An array is used instead of a slice because arrays are not
// passed by reference.
type Sudoku struct {
	cells [81]square
}

// The two different types of squares do not share methods, so we are using the
// empty interface here.  Anytime the algorithm uses the fields it needs to do
// a typeswitch anyway.
type square interface{}

// A filled out square is just the value it represents. Uint8 is sufficient, we
// are only storing values from 1 to 9.
type filledOutSquare uint8

// An empty square is more interesting. In addition to its emptiness (encoded
// by the type) it also contains the information regarding already eliminated
// values. If a value is eliminated i.e. it can't occur in this cell, its
// corresponding bit is set in the eliminatedValues field. As it is uint16,
// there is plenty of space for 9 different values.  Additionally, we cache the
// number of eliminated values, since this is used later on.
//
// We are using a bitset instead of a possible map[uint]bool for efficiency (hopefully), but
// mostly because uint16 values are passed by value.
type emptySquare struct {
	eliminatedValues         uint16
	numberOfEliminatedValues uint8
}

// This method eliminates one possible value from this square. If there is only
// one value left, it returns a filled square with this value, else it returns
// the old square minus the given value.
func (es emptySquare) eliminated(sv uint8) square {
	if es.eliminatedValues&(1<<sv) == 0 {
		es.numberOfEliminatedValues++
		es.eliminatedValues |= 1 << sv
	}

	if es.numberOfEliminatedValues == 8 {
		return filledOutSquare(es.possibleValues()[0])
	}

	return es
}

func (es emptySquare) possibleValues() []uint8 {
	var res []uint8
	for i := uint8(1); i <= uint8(9); i++ {
		if es.isValuePossible(i) {
			res = append(res, i)
		}
	}
	return res
}

func (es emptySquare) isValuePossible(sv uint8) bool {
	return es.eliminatedValues&(1<<sv) == 0
}

// Parsing

func parseCell(r, c rune, sudoku Sudoku, rr io.RuneReader) (Sudoku, error) {
	var x rune
	var err error

	for x = ' '; x != '.' && (x < '0' || x > '9'); x, _, err = rr.ReadRune() {
		if err != nil {
			return sudoku, err
		}
	}

	if x >= '1' && x <= '9' {
		return sudoku.withAssignment(coord(r, c), uint8(x-'0'))
	}
	return sudoku, nil
}

// ParseReader reads a complete sudoku from the given rune reader. The
// following semantics apply:
//
// * Any digit except zero fills the cell directly. If a conflict arises (same
// number in same column, for example), an error is returned.
//
// * A zero or dot (0 or .) are interpreted as empty field.
//
// * Any other rune is ignored.
//
// Thanks to this it is possible to parse a sudoku in complex format as well as
// in a single row.
func ParseReader(rr io.RuneReader) (Sudoku, error) {
	var sudoku Sudoku

	for ind := range sudoku.cells {
		sudoku.cells[ind] = emptySquare{}
	}

	for r := 'A'; r <= 'I'; r++ {
		for c := '1'; c <= '9'; c++ {
			var err error
			if sudoku, err = parseCell(r, c, sudoku, rr); err != nil {
				return sudoku, err
			}
		}
	}

	return sudoku, nil
}

// Parse is a convenience wrapper for ParseReader that accepts a string. See
// ParseReader for details.
func Parse(s string) (Sudoku, error) {
	return ParseReader(strings.NewReader(s))
}

// Solve tries to solve the receiver by trial-and-error, filling in one field and
// continuing until a conflict occurs. If there is no solution, an error is returned,
// otherwise the solution is returned.
//
// If there are multiple solutions to a sudoku, i.e. it's underspecified, one
// of them is returned.
func (s Sudoku) Solve() (Sudoku, error) {
	var coordWithMaximumEliminatedValues coordinate
	maximumEliminatedValues := uint8(0)
	solved := true
	for coord, x := range s.cells {
		// accept zero state as empty square
		if x == nil {
			s.cells[coord] = emptySquare{}
			x = s.cells[coord]
		}
		if val, ok := x.(emptySquare); ok {
			solved = false

			if val.numberOfEliminatedValues >= maximumEliminatedValues {
				maximumEliminatedValues = val.numberOfEliminatedValues
				coordWithMaximumEliminatedValues = coordinate(coord)
			}
		}
	}
	if solved {
		return s, nil
	}

	for _, sv := range s.cells[coordWithMaximumEliminatedValues].(emptySquare).possibleValues() {
		news, err := s.withAssignment(coordWithMaximumEliminatedValues, sv)
		if err != nil {
			continue
		}

		if solved, err := news.Solve(); err == nil {
			return solved, nil
		}
	}
	return s, ErrConflict
}

// WithCellValued returns a new sudoku with the field at position rc filled in
// with the given value.  If a conflict arises due to this assignment, an error
// is returned.
func (s Sudoku) WithCellValued(r, c rune, sv uint8) (Sudoku, error) {
	return s.withAssignment(coord(r, c), sv)
}

func (s Sudoku) withAssignment(c coordinate, sv uint8) (Sudoku, error) {
	if es, ok := s.cells[c].(emptySquare); ok && !es.isValuePossible(sv) {
		// field is empty, but can't take that value
		return s, ErrConflict
	}
	s.cells[c] = filledOutSquare(sv)

	for peerC := range peers[c] {
		peer := s.cells[peerC]

		switch sq := peer.(type) {
		case filledOutSquare:
			if uint8(sq) == sv {
				// conflict, we are asked to remove the value we already have
				return s, ErrConflict
			}
		case emptySquare:
			newsq := sq.eliminated(sv)
			if fos, ok := newsq.(filledOutSquare); ok {
				// Propagate
				var err error
				if s, err = s.withAssignment(peerC, uint8(fos)); err != nil {
					return s, err
				}
			} else {
				// just assign the changed field
				s.cells[peerC] = newsq
			}
		}
	}
	return s, nil
}

// Output

// AsInts returns the receiver as a 9x9 grid suitable for display. Any
// non-filled cells are returned as zero (0). The returned grid is not
// connected to the internal data structures and may be modified freely.
func (s Sudoku) AsInts() [9][9]uint8 {
	var res [9][9]uint8
	for i := range s.cells {
		var cellValue uint8

		switch sq := s.cells[i].(type) {
		case filledOutSquare:
			cellValue = uint8(sq)
		default:
			cellValue = 0
		}
		res[i/9][i%9] = cellValue
	}
	return res
}

// String gives the underlying sudoku as a string, with lines separating the
// blocks.  See the examples for the structure.
func (s Sudoku) String() string {
	var res string

	for r := 'A'; r <= 'I'; r++ {
		for c := '1'; c <= '9'; c++ {
			switch sq := s.cells[coord(r, c)].(type) {
			case filledOutSquare:
				res += fmt.Sprintf("%v", uint8(sq))
			default:
				res += "."
			}
			switch {
			case c == '9':
				res += "\n"
			case (c-'0')%3 == 0:
				res += " |"
			default:
				res += " "
			}
		}
		if r == 'C' || r == 'F' {
			res += "------+------+------\n"
		}
	}
	return res
}

// Coordinates are represented by bytes, they are the indices in the cells
// array.
type coordinate uint8

func coord(r, c rune) coordinate {
	return coordinate(uint8(r-'A')*9 + uint8(c-'1'))
}

// Peers Calculation

// A peer is any cell that is influenced by the key, for example A1 is peer of
// A2, A3, B1, B3 etc, but not of D9.
var peers = make(map[coordinate]map[coordinate]struct{})

func addPeersFor(r, c rune) {
	cr := coord(r, c)

	peers[cr] = make(map[coordinate]struct{})
	for r2 := 'A'; r2 <= 'I'; r2++ {
		if r2 != r {
			peers[cr][coord(r2, c)] = struct{}{}
		}
	}
	for c2 := '1'; c2 <= '9'; c2++ {
		if c2 != c {
			peers[cr][coord(r, c2)] = struct{}{}
		}
	}
	rowOffset := (r - 'A') % 3
	colOffset := (c - '1') % 3
	for r2 := r - rowOffset; r2 <= r-rowOffset+2; r2++ {
		for c2 := c - colOffset; c2 <= c-colOffset+2; c2++ {
			if r2 != r && c2 != c {
				peers[cr][coord(r2, c2)] = struct{}{}
			}
		}
	}
}

func init() {
	for r := 'A'; r <= 'I'; r++ {
		for c := '1'; c <= '9'; c++ {
			addPeersFor(r, c)
		}
	}
}
