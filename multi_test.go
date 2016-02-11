package sudoku

import (
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

type durationsSorter struct {
	vals []time.Duration
}

func (d durationsSorter) Len() int {
	return len(d.vals)
}

func (d durationsSorter) Less(i, j int) bool {
	return d.vals[i] < d.vals[j]
}

func (d durationsSorter) Swap(i, j int) {
	d.vals[i], d.vals[j] = d.vals[j], d.vals[i]
}

func testAllIn(filename string, t *testing.T) {
	contents, err := ioutil.ReadFile(filepath.Join("fixtures", filename))
	if err != nil {
		t.Fatal(err)
	}

	rr := strings.NewReader(string(contents))

	var times []time.Duration

	for {
		timeStart := time.Now()
		sudoku, err := ParseReader(rr)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Error(err)
		}
		solution, err := sudoku.Solve()
		timeEnd := time.Now()

		if err != nil {
			t.Error(err)
			continue
		}
		assertIsValidSudoku(solution, t)

		times = append(times, timeEnd.Sub(timeStart))
	}

	var max, min, sum, count int64
	min = math.MaxInt64

	for _, x := range times {
		n := x.Nanoseconds()
		if n > max {
			max = n
		}
		if n < min {
			min = n
		}
		sum += n
		count++
	}

	sorter := durationsSorter{vals: times}
	sort.Sort(sorter)

	var median time.Duration
	if count > 0 {
		median = times[count/2]
	}

	fmt.Printf("%s: min=%v max=%v avg=%v median=%v (%v sudokus in %v)\n", filename, time.Duration(min), time.Duration(max), median, time.Duration(sum/count), count, time.Duration(sum))
}

func TestEasy(t *testing.T) {
	testAllIn("easy50.txt", t)
}
func TestHardest(t *testing.T) {
	testAllIn("hardest.txt", t)
}
func TestTop95(t *testing.T) {
	if testing.Short() {
		t.Skip("Top 95 takes seconds")
	}
	testAllIn("top95.txt", t)
}
