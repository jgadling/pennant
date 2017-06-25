package main

import (
	"bytes"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_ColumnPrinter(t *testing.T) {
	colTests := []struct {
		name      string
		separator string
		headings  []string
		rows      [][]string
		result    string
	}{
		{
			"single column", " ",
			[]string{"one"},
			[][]string{{"two"}, {"three"}},
			"one  \n-----\ntwo  \nthree\n",
		}, {
			"two columns table", " ",
			[]string{"one", "two"},
			[][]string{{"three", "four"}, {"five"}},
			"one   two \n----- ----\nthree four\nfive \n",
		}, {
			"too many columns", " ",
			[]string{"many", "cols"},
			[][]string{{"three", "four", "five"}, {"six"}},
			"many  cols\n----- ----\nthree four\nsix  \n",
		}, {
			"no rows", " ",
			[]string{"no", "rows"},
			[][]string{},
			"no rows\n-- ----\n",
		},
	}
	for _, testCase := range colTests {
		testDescription := fmt.Sprintf("The output for %s should match", testCase.name)
		Convey(testDescription, t, func() {
			var buffer bytes.Buffer
			cp := NewColPrinter(testCase.headings, testCase.separator, &buffer)
			for _, row := range testCase.rows {
				cp.AddRow(row)
			}
			cp.Print()
			So(buffer.String(), ShouldEqual, testCase.result)
		})
	}
}
