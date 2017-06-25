package main

import (
	"fmt"
	"io"
	"strings"
)

// ColumnPrinter represents a table of data to be pretty-printed
type ColumnPrinter struct {
    output    io.Writer
	headings  []string
	widths    []int
	separator string
	data      [][]string
}

// Initialize a new ColumnPrinter with heading labels and the separator to use
// between columns
func NewColPrinter(headings []string, separator string, file io.Writer) *ColumnPrinter {
	widths := make([]int, len(headings))
	for i, heading := range headings {
		widths[i] = len(heading)
	}
	cp := ColumnPrinter{
		output:    file,
		headings:  headings,
		widths:    widths,
		separator: separator}
	return &cp
}

// Add a row of data to the table
func (cp *ColumnPrinter) AddRow(row []string) {
	for i, cell := range row {
		if i >= len(cp.headings) {
			break
		}
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
	colCount := len(cp.headings)
	if len(row) < colCount {
		colCount = len(row)
	}
	for i, cell := range row {
		if i >= colCount {
			break
		}
		row[i] = rpad(cell, cp.widths[i], padding)
	}
	fmt.Fprintln(cp.output, strings.Join(row[:colCount], cp.separator))
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
