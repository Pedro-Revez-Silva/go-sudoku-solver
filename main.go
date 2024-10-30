package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"
)

const (
	SIZE      = 9
	EMPTY     = '.'
	GRID_SIZE = SIZE * SIZE
	ALL_BITS  = 0x1FF
)

// Pre-calculated lookup tables
var (
	rowMasks    [SIZE][SIZE]uint16
	colMasks    [SIZE][SIZE]uint16
	boxMasks    [SIZE][SIZE]uint16
	bitCount    [512]int
	firstDigit  [512]int
	digitValues [9]byte
)

func init() {
	for i := 0; i < 9; i++ {
		digitValues[i] = byte(i + 1)
	}

	for i := 0; i < 512; i++ {
		count := 0
		for j := 0; j < 9; j++ {
			if i&(1<<j) != 0 {
				count++
			}
		}
		bitCount[i] = count
	}

	for i := 0; i < 512; i++ {
		firstDigit[i] = -1
		for j := 0; j < 9; j++ {
			if i&(1<<j) != 0 {
				firstDigit[i] = j
				break
			}
		}
	}

	for i := 0; i < SIZE; i++ {
		for j := 0; j < SIZE; j++ {
			rowMasks[i][j] = uint16(1 << j)
			colMasks[i][j] = uint16(1 << i)
			boxMasks[i][j] = uint16(1 << ((i/3)*3 + j/3))
		}
	}
}

type Puzzle struct {
	cells     [SIZE][SIZE]byte
	rows      [SIZE]uint16
	cols      [SIZE]uint16
	boxes     [SIZE]uint16
	emptyCell int
}

func ParsePuzzle(input string) *Puzzle {
	p := &Puzzle{}
	idx := 0
	for i := 0; i < SIZE; i++ {
		for j := 0; j < SIZE; j++ {
			if input[idx] != EMPTY {
				digit := input[idx] - '1'
				p.cells[i][j] = digit + 1
				p.rows[i] |= 1 << digit
				p.cols[j] |= 1 << digit
				p.boxes[(i/3)*3+j/3] |= 1 << digit
			} else {
				p.emptyCell++
			}
			idx++
		}
	}
	return p
}

func getBox(row, col int) int {
	return (row/3)*3 + col/3
}

func (p *Puzzle) getPossibilities(row, col int) uint16 {
	box := getBox(row, col)
	return ^(p.rows[row] | p.cols[col] | p.boxes[box]) & ALL_BITS
}

func (p *Puzzle) findBestCell() (int, int, uint16, bool) {
	if p.emptyCell == 0 {
		return 0, 0, 0, false
	}

	minRow, minCol := 0, 0
	minPoss := uint16(ALL_BITS)
	minCount := 10

	for i := 0; i < SIZE; i++ {
		for j := 0; j < SIZE; j++ {
			if p.cells[i][j] == 0 {
				poss := p.getPossibilities(i, j)
				count := bitCount[poss]
				if count < minCount {
					minCount = count
					minPoss = poss
					minRow = i
					minCol = j
					if count == 1 {
						return minRow, minCol, minPoss, true
					}
				}
			}
		}
	}
	return minRow, minCol, minPoss, true
}

func (p *Puzzle) setCell(row, col int, val byte) {
	p.cells[row][col] = val
	bit := uint16(1 << (val - 1))
	p.rows[row] |= bit
	p.cols[col] |= bit
	p.boxes[getBox(row, col)] |= bit
	p.emptyCell--
}

func (p *Puzzle) clearCell(row, col int, val byte) {
	p.cells[row][col] = 0
	bit := ^uint16(1 << (val - 1))
	p.rows[row] &= bit
	p.cols[col] &= bit
	p.boxes[getBox(row, col)] &= bit
	p.emptyCell++
}

func (p *Puzzle) solve() bool {
	row, col, poss, found := p.findBestCell()
	if !found {
		return true
	}

	for poss != 0 {
		digit := uint16(firstDigit[poss] + 1)
		val := byte(digit)
		p.setCell(row, col, val)

		if p.solve() {
			return true
		}
		p.clearCell(row, col, val)
		poss &= ^(1 << (digit - 1))
	}
	return false
}

func (p *Puzzle) ToString() string {
	result := make([]byte, GRID_SIZE)
	idx := 0
	for i := 0; i < SIZE; i++ {
		for j := 0; j < SIZE; j++ {
			if p.cells[i][j] == 0 {
				result[idx] = EMPTY
			} else {
				result[idx] = p.cells[i][j] + '0'
			}
			idx++
		}
	}
	return string(result)
}

func solvePuzzles(puzzles []string) []string {
	numWorkers := runtime.NumCPU()

	jobs := make(chan int, len(puzzles))
	results := make(chan struct {
		index    int
		solution string
	}, len(puzzles))

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				puzzle := ParsePuzzle(puzzles[idx])
				if puzzle.solve() {
					results <- struct {
						index    int
						solution string
					}{idx, puzzle.ToString()}
				}
			}
		}()
	}

	for i := range puzzles {
		jobs <- i
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	solutions := make([]string, len(puzzles))
	for result := range results {
		solutions[result.index] = result.solution
	}

	return solutions
}

func main() {
	var puzzles []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == GRID_SIZE {
			puzzles = append(puzzles, line)
		}
	}

	start := time.Now()
	solutions := solvePuzzles(puzzles)
	duration := time.Since(start)
	fmt.Printf("Solved %d puzzles in %v\n", len(puzzles), duration)
	fmt.Printf("Average time per puzzle: %v\n", duration/time.Duration(len(puzzles)))

	// Write solutions
	file, _ := os.Create("solutions.txt")
	writer := bufio.NewWriter(file)
	for _, solution := range solutions {
		writer.WriteString(solution + "\n")
	}
	writer.Flush()
	defer file.Close()
}
