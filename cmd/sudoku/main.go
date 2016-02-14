//+build !appengine

// This is an example for using the sudoku package. It reads one sudoku from
// stdin and prints out the solution, if any. Otherwise, it prints a message
// and exits with code 1.
package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/thriqon/sudoku"
)

func main() {
	sudoku, err := sudoku.ParseReader(bufio.NewReader(os.Stdin))
	if err != nil {
		fmt.Println(err)
		return
	}

	solved, err := sudoku.Solve()

	if err != nil {
		fmt.Println("NO SOLUTION FOUND")
		os.Exit(1)
		return
	}

	fmt.Print(solved.String())
}
