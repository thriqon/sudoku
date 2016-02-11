package sudoku

import (
	"fmt"
	"sort"
	"testing"
)

func TestSquare(t *testing.T) {
	var s square
	s = emptySquare{}

	if s.(emptySquare).eliminatedValues != 0 {
		t.Error("Expected no eliminated values")
	}

	s = s.(emptySquare).eliminated(1)

	if s.(emptySquare).eliminatedValues != (1 << 1) {
		t.Error("Expected to have eliminated 1, but was", s.(emptySquare).eliminated)
	}

	for i := uint8(1); i <= uint8(8); i++ {
		s = s.(emptySquare).eliminated(i)
	}

	if val := s.(filledOutSquare); val != 9 {
		t.Error("Should have value = 9, but was", val)
	}
}

func TestGlobals(t *testing.T) {
	if len(peers[coord('A', '2')]) != 20 {
		t.Error("Not 20 peers")
	}
}

func TestSudokuString(t *testing.T) {
	txt := `4 . . |. . . |8 . 5
. 3 . |. . . |. . .
. . . |7 . . |. . .
------+------+------
. 2 . |. . . |. 6 .
. . . |. 8 . |4 . .
. . . |. 1 . |. . .
------+------+------
. . . |6 . 3 |. 7 .
5 . . |2 . . |. . .
1 . 4 |. . . |. . .
`
	sudoku, err := Parse(txt)
	if err != nil {
		t.Fatal(err)
	}
	if val := sudoku.cells[coord('A', '1')].(filledOutSquare); val != 4 {
		t.Error("Expected A1:4, but ", val)
	}

	txted := sudoku.String()
	if txted != txt {
		t.Error("expected \n", txt, "but \n", txted)
	}
}

func ExampleSudoku() {
	// a Sudoku needs no initialization
	s := Sudoku{}
	_, err := s.Solve()
	if err != nil {
		panic(err)
	}
}

func ExampleSudoku_Solve_hardestByInkala1() {
	source := `8 5 . |. . 2 |4 . .
	7 2 . |. . . |. . 9
	. . 4 |. . . |. . .
	------+------+------
	. . . |1 . 7 |. . 2
	3 . 5 |. . . |9 . .
	. 4 . |. . . |. . .
	------+------+------
	. . . |. 8 . |. 7 .
	. 1 7 |. . . |. . .
	. . . |. 3 6 |. 4 .
	`

	parsed, err := Parse(source)
	if err != nil {
		panic(err)
	}
	solution, err := parsed.Solve()
	if err != nil {
		panic(err)
	}
	fmt.Println(solution)
	// Output:
	// 8 5 9 |6 1 2 |4 3 7
	// 7 2 3 |8 5 4 |1 6 9
	// 1 6 4 |3 7 9 |5 2 8
	// ------+------+------
	// 9 8 6 |1 4 7 |3 5 2
	// 3 7 5 |2 6 8 |9 1 4
	// 2 4 1 |5 9 3 |7 8 6
	// ------+------+------
	// 4 3 2 |9 8 1 |6 7 5
	// 6 1 7 |4 2 5 |8 9 3
	// 5 9 8 |7 3 6 |2 4 1
}

func ExampleSudoku_Solve_hardestByInkala2() {
	source := `. . 5 |3 . . |. . .
	8 . . |. . . |. 2 .
	. 7 . |. 1 . |5 . .
	------+------+------
	4 . . |. . 5 |3 . .
	. 1 . |. 7 . |. . 6
	. . 3 |2 . . |. 8 .
	------+------+------
	. 6 . |5 . . |. . 9
	. . 4 |. . . |. 3 .
	. . . |. . 9 |7 . .
	`

	parsed, err := Parse(source)
	if err != nil {
		panic(err)
	}
	solution, err := parsed.Solve()
	if err != nil {
		panic(err)
	}
	fmt.Println(solution)
	// Output:
	// 1 4 5 |3 2 7 |6 9 8
	// 8 3 9 |6 5 4 |1 2 7
	// 6 7 2 |9 1 8 |5 4 3
	// ------+------+------
	// 4 9 6 |1 8 5 |3 7 2
	// 2 1 8 |4 7 3 |9 5 6
	// 7 5 3 |2 9 6 |4 8 1
	// ------+------+------
	// 3 6 7 |5 4 2 |8 1 9
	// 9 8 4 |7 6 1 |2 3 5
	// 5 2 1 |8 3 9 |7 6 4
}

func assertIsValidUnit(unit [9]uint8, t *testing.T) {
	unitI := make([]int, 9)
	for ind := range unit {
		unitI[ind] = int(unit[ind])
	}
	sort.Ints(unitI)

	if actual := fmt.Sprint(unitI); actual != "[1 2 3 4 5 6 7 8 9]" {
		t.Error("Unit does not contain exactly the expected values:", actual)
	}
}

func assertIsValidSudoku(s Sudoku, t *testing.T) {
	cells := s.AsInts()

	for _, row := range cells {
		assertIsValidUnit(row, t)
	}

	for c := 0; c < 9; c++ {
		col := [9]uint8{
			cells[0][c], cells[1][c], cells[2][c],
			cells[3][c], cells[4][c], cells[5][c],
			cells[6][c], cells[7][c], cells[8][c],
		}
		assertIsValidUnit(col, t)
	}

	for x := 0; x < 3; x++ {
		for y := 0; y < 3; y++ {
			box := [9]uint8{
				cells[x*3][y*3], cells[x*3+1][y*3], cells[x*3+2][y*3],
				cells[x*3][y*3+1], cells[x*3+1][y*3+1], cells[x*3+2][y*3+1],
				cells[x*3][y*3+2], cells[x*3+1][y*3+2], cells[x*3+2][y*3+2],
			}
			assertIsValidUnit(box, t)
		}
	}
}

func TestRejectsInvalidSudoku(t *testing.T) {
	txt := `4 . 4 |. . . |8 . 5
. 3 . |. . . |. . .
. . . |7 . . |. . .
------+------+------
. 2 . |. . . |. 6 .
. . . |. 8 . |4 . .
. . . |. 1 . |. . .
------+------+------
. . . |6 . 3 |. 7 .
5 . . |2 . . |. . .
1 . 4 |. . . |. . .
`
	_, err := Parse(txt)
	if err == nil {
		t.Fatal("Unexpected non-error")
	}
}

func TestIncompleteSudokuAsUints(t *testing.T) {
	txt := `4 . . |. . . |8 . 5
. 3 . |. . . |. . .
. . . |7 . . |. . .
------+------+------
. 2 . |. . . |. 6 .
. . . |. 8 . |4 . .
. . . |. 1 . |. . .
------+------+------
. . . |6 . 3 |. 7 .
5 . . |2 . . |. . .
1 . 4 |. . . |. . .
`
	sudoku, err := Parse(txt)
	if err != nil {
		t.Fatal(sudoku)
	}
	expected := [9][9]uint8{
		{4, 0, 0, 0, 0, 0, 8, 0, 5},
		{0, 3, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 7, 0, 0, 0, 0, 0},
		{0, 2, 0, 0, 0, 0, 0, 6, 0},
		{0, 0, 0, 0, 8, 0, 4, 0, 0},
		{0, 0, 0, 0, 1, 0, 0, 0, 0},
		{0, 0, 0, 6, 0, 3, 0, 7, 0},
		{5, 0, 0, 2, 0, 0, 0, 0, 0},
		{1, 0, 4, 0, 0, 0, 0, 0, 0},
	}

	if actual := sudoku.AsInts(); actual != expected {
		t.Error("Actual is not expected: ", actual)
	}
}

func TestFindsSolutionFromEmptySudoku(t *testing.T) {
	s := Sudoku{}
	solved, err := s.Solve()
	if err != nil {
		t.Fatal("Did not find solution")
	}
	assertIsValidSudoku(solved, t)
}

func TestFindsSolutionFromAPISetSudoku(t *testing.T) {
	s, err := (Sudoku{}).WithCellValued('A', '1', 5)
	if err != nil {
		t.Fatal(err)
	}

	solved, err := s.Solve()
	if err != nil {
		t.Fatal("No solution found")
	}
	assertIsValidSudoku(solved, t)
}

func TestRejectsConflictsViaAPI(t *testing.T) {
	s, err := (Sudoku{}).WithCellValued('A', '1', 5)
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.WithCellValued('A', '2', 5)
	if err == nil {
		t.Error("Should return error")
	}
}
