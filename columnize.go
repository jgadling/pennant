package main

import (
	"fmt"
	"strings"
)

// ColumnPrinter represents a table of data to be pretty-printed
type ColumnPrinter struct {
	headings  []string
	widths    []int
	separator string
	data      [][]string
}

// Initialize a new ColumnPrinter with heading labels and the separator to use
// between columns
func NewColPrinter(headings []string, separator string) *ColumnPrinter {
	widths := make([]int, len(headings))
	for i, heading := range headings {
		widths[i] = len(heading)
	}
	cp := ColumnPrinter{
		headings:  headings,
		widths:    widths,
		separator: separator}
	return &cp
}

// Add a row of data to the table
func (cp *ColumnPrinter) AddRow(row []string) {
	for i, cell := range row {
		if len(cell) > cp.widths[i] {
			cp.widths[i] = len(cell)
		}
	}
	cp.data = append(cp.data, row)
}

// Right-pad a string to a given width, with a padding character.
func rpad(str string, width int, padding string) string {
	for true {
		if len(str) >= width {
			return str
		}
		str += padding
	}
	return str
}

// Format a single row and print it.
func (cp *ColumnPrinter) printRow(row []string, padding string) {
	for i, cell := range row {
		row[i] = rpad(cell, cp.widths[i], padding)
	}
	fmt.Println(strings.Join(row, cp.separator))
}

// Pretty-print the table.
func (cp *ColumnPrinter) Print() {
	cp.printRow(cp.headings, " ")
	separators := make([]string, len(cp.headings))
	cp.printRow(separators, "-")
	for _, row := range cp.data {
		cp.printRow(row, " ")
	}
}
