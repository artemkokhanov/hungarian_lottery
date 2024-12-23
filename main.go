package main

import (
	"bufio"
	"fmt"
	"math/bits"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	MAX_NUM         = 90 // Range of possible numbers to chose from
	BITS_PER_UINT64 = 64 // The number of bits to store per uint64
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: go main.go input_file")
		os.Exit(1)
	}
	inputFilePath := os.Args[1]

	// STEP 1 - PREPROCESSING:
	totalPlayers := countFileLines(inputFilePath)

	bitsetLength := (totalPlayers + BITS_PER_UINT64 - 1) / BITS_PER_UINT64 // To avoid flooring

	chosenNumbersBitsets := make([][]uint64, MAX_NUM+1) // +1 to start indexing at 1 for convenience
	for i := 1; i <= MAX_NUM; i++ {
		chosenNumbersBitsets[i] = make([]uint64, bitsetLength)
	}

	populateBitsets(inputFilePath, chosenNumbersBitsets, totalPlayers)

	fmt.Println("READY")

	// STEP 2 - QUERYING:
	// Read queries from standard input until EOF
	queryScanner := bufio.NewScanner(os.Stdin)
	for queryScanner.Scan() {
		line := queryScanner.Text()
		parts := strings.Fields(line)
		if len(parts) != 5 { // Skip the query if it doesn't have exactly 5 numbers
			continue
		}

		start := time.Now() // Record the start time to measure how long the query takes

		// Parse the query numbers
		selectedWinningNumbers := make([]int, 5)
		isValid := true
		for i := 0; i < 5; i++ {
			val, err := strconv.Atoi(parts[i])
			if err != nil || val < 1 || val > MAX_NUM { // Number must be between 1 and MAX_NUM
				isValid = false
				break
			}
			selectedWinningNumbers[i] = val
		}
		if !isValid {
			continue // Skip if invalid
		}

		countAtLeastTwo, countAtLeastThree, countAtLeastFour, countAllFive := calculateIntersections(chosenNumbersBitsets, selectedWinningNumbers)

		// Use inclusion-exclusion to compute matches:
		exactly5 := countAllFive
		exactly4 := countAtLeastFour - 5*exactly5
		exactly3 := countAtLeastThree - 4*exactly4 - 10*exactly5
		exactly2 := countAtLeastTwo - 3*exactly3 - 6*exactly4 - 10*exactly5

		fmt.Printf("%d %d %d %d\n", exactly2, exactly3, exactly4, exactly5) // Print the result

		elapsed := time.Since(start)
		fmt.Fprintf(os.Stderr, "Query took: %d ms\n", elapsed.Milliseconds())
	}
}

// HELPER METHODS:

func countFileLines(filename string) int {
	// Returns the number of lines (players) in the input file.
	f, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lines := 0
	for scanner.Scan() {
		lines++
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}
	return lines
}

func populateBitsets(filename string, chosenNumbersBitsets [][]uint64, totalPlayers int) {
	// Reads the file and fills the bitsets datastructure
	// For player i, if they chose number n, we set the i-th bit in chosenNumbersBitsets[n]
	f, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	currentPlayerIndex := 0
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) != 5 {
			fmt.Fprintf(os.Stderr, "Warning: Invalid line format at player index %d, no numbers set for this player.\n", currentPlayerIndex)
			currentPlayerIndex++
			continue
		}

		chosenNumbers := make([]int, 5)
		isValid := true
		for i := range chosenNumbers {
			val, err := strconv.Atoi(parts[i])
			if err != nil || val < 1 || val > MAX_NUM {
				// Invalid number: warn and do not set any bits for this player
				isValid = false
				break
			}
			chosenNumbers[i] = val
		}

		if !isValid {
			fmt.Fprintf(os.Stderr, "Warning: Invalid number in line at player index %d, no numbers set for this player.\n", currentPlayerIndex)
			currentPlayerIndex++
			continue
		}

		// Calculate the position of this player's bit in the array of uint64
		wordIndex := currentPlayerIndex / BITS_PER_UINT64
		bitOffset := uint(currentPlayerIndex % BITS_PER_UINT64)
		mask := uint64(1) << bitOffset

		// Set the bits for each chosen number
		for _, n := range chosenNumbers {
			chosenNumbersBitsets[n][wordIndex] |= mask
		}
		currentPlayerIndex++
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	if currentPlayerIndex != totalPlayers {
		fmt.Fprintf(os.Stderr, "Line count mismatch. Expected %d got %d\n", totalPlayers, currentPlayerIndex)
		os.Exit(1)
	}
}

func calculateIntersections(chosenNumbersBitsets [][]uint64, selectedWinningNumbers []int) (countAtLeastTwo, countAtLeastThree, countAtLeastFour, countAllFive int) {
	winningNumberBitsets := make([][]uint64, 5)
	for i := 0; i < 5; i++ {
		winningNumberBitsets[i] = chosenNumbersBitsets[selectedWinningNumbers[i]]
	}

	// Count players in each pairwise intersection
	pairs := [][2]int{
		{0, 1}, {0, 2}, {0, 3}, {0, 4},
		{1, 2}, {1, 3}, {1, 4},
		{2, 3}, {2, 4},
		{3, 4},
	}
	for _, p := range pairs {
		countAtLeastTwo += countSetBitsAND(winningNumberBitsets[p[0]], winningNumberBitsets[p[1]])
	}

	// Count players in each triple intersection
	triples := [][3]int{
		{0, 1, 2}, {0, 1, 3}, {0, 1, 4}, {0, 2, 3}, {0, 2, 4},
		{0, 3, 4}, {1, 2, 3}, {1, 2, 4}, {1, 3, 4}, {2, 3, 4},
	}
	for _, t := range triples {
		countAtLeastThree += countSetBitsAND3(winningNumberBitsets[t[0]], winningNumberBitsets[t[1]], winningNumberBitsets[t[2]])
	}

	// Count players in each quadruple intersection
	quadruples := [][4]int{
		{0, 1, 2, 3}, {0, 1, 2, 4}, {0, 1, 3, 4}, {0, 2, 3, 4}, {1, 2, 3, 4},
	}
	for _, q := range quadruples {
		countAtLeastFour += countSetBitsAND4(winningNumberBitsets[q[0]], winningNumberBitsets[q[1]], winningNumberBitsets[q[2]], winningNumberBitsets[q[3]])
	}

	// Count players that chose all 5 numbers
	countAllFive = countSetBitsAND5(winningNumberBitsets[0], winningNumberBitsets[1], winningNumberBitsets[2], winningNumberBitsets[3], winningNumberBitsets[4])

	return
}

func countSetBits(a []uint64) int {
	// Counts how many bits are set in a bitset.
	count := 0
	for _, word := range a {
		count += bits.OnesCount64(word)
	}
	return count
}

func countSetBitsAND(a, b []uint64) int {
	count := 0
	for i := range a {
		count += bits.OnesCount64(a[i] & b[i])
	}
	return count
}

func countSetBitsAND3(a, b, c []uint64) int {
	count := 0
	for i := range a {
		count += bits.OnesCount64(a[i] & b[i] & c[i])
	}
	return count
}

func countSetBitsAND4(a, b, c, d []uint64) int {
	count := 0
	for i := range a {
		count += bits.OnesCount64(a[i] & b[i] & c[i] & d[i])
	}
	return count
}

func countSetBitsAND5(a, b, c, d, e []uint64) int {
	count := 0
	for i := range a {
		count += bits.OnesCount64(a[i] & b[i] & c[i] & d[i] & e[i])
	}
	return count
}
