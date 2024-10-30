package main

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

const (
	SIZE      = 9
	EMPTY     = '.'
	GRID_SIZE = SIZE * SIZE
	ALL_BITS  = 0x1FF
)

var (
	bitCount   [512]int
	firstDigit [512]int
)

func init() {
	for i := 0; i < 512; i++ {
		count := 0
		first := -1
		for j := 0; j < 9; j++ {
			if i&(1<<j) != 0 {
				if first == -1 {
					first = j
				}
				count++
			}
		}
		bitCount[i] = count
		firstDigit[i] = first
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

func main() {
	var puzzles []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == GRID_SIZE {
			puzzles = append(puzzles, line)
		}
	}

	file, _ := os.Create("solutions.txt")
	writer := bufio.NewWriter(file)
	defer file.Close()

	start := time.Now()
	solved := 0

	for _, puzzleStr := range puzzles {
		puzzle := ParsePuzzle(puzzleStr)
		if puzzle.solve() {
			writer.WriteString(puzzle.ToString() + "\n")
			solved++
		} else {
			writer.WriteString("No solution found\n")
		}
	}

	writer.Flush()

	duration := time.Since(start)
	fmt.Printf("Solved %d puzzles in %v\n", solved, duration)
	fmt.Printf("Average time per puzzle: %v\n", duration/time.Duration(len(puzzles)))
}
