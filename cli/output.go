package cli

import "fmt"

func renderTable(data [][]string) {
	if len(data) == 0 {
		return
	}

	// Find column widths
	cols := len(data[0])
	widths := make([]int, cols)

	for _, row := range data {
		for i, cell := range row {
			if len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	for _, row := range data {
		for i, cell := range row {
			fmt.Printf("  %-*s", widths[i]+2, cell)
		}

		fmt.Println()
	}
}
