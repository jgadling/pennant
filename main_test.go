package main

import (
	"bytes"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestCli(t *testing.T) {
	Convey("Test returns false", t, func() {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		runCli([]string{
			"placeholder", "test", "-f", "tests/data/flag1.json", "-d",
			"tests/data/data2.json"}, &stdout, &stderr)
		So(stdout.String(), ShouldEqual, "Flag        Status \n----------  -------\nred_button  enabled\n")
	})
	Convey("Test returns true", t, func() {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		runCli([]string{
			"placeholder", "test", "-f", "tests/data/flag1.json", "-d",
			"tests/data/data1.json"}, &stdout, &stderr)
		So(stdout.String(), ShouldEqual, "Flag        Status  \n----------  --------\nred_button  disabled\n")
	})
}
