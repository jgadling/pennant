package main

import (
	"fmt"
	"strings"
)

type ColumnPrinter struct {
	headings  []string
	widths    []int
	separator string
	data      [][]string
}

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

func (cp *ColumnPrinter) AddRow(row []string) {
	for i, cell := range row {
		if len(cell) > cp.widths[i] {
			cp.widths[i] = len(cell)
		}
	}
	cp.data = append(cp.data, row)
}

func rpad(str string, width int, padding string) string {
	for true {
		if len(str) >= width {
			return str
		}
		str += padding
	}
	return str
}

func (cp *ColumnPrinter) printRow(row []string, padding string) {
	for i, cell := range row {
		row[i] = rpad(cell, cp.widths[i], padding)
	}
	fmt.Println(strings.Join(row, cp.separator))
}

func (cp *ColumnPrinter) Print() {
	cp.printRow(cp.headings, " ")
	separators := make([]string, len(cp.headings))
	cp.printRow(separators, "-")
	for _, row := range cp.data {
		cp.printRow(row, " ")
	}
}
