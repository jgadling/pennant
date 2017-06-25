package main

import (
	"os"
	"os/exec"
	"flag"
	"testing"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	testIntegConsul = flag.Bool("integ-consul", false, "run consul integration tests")
	testIntegStatsd = flag.Bool("integ-statsd", false, "run statsd integration tests")
	testIntegration = flag.Bool("integration", false, "run all integration tests")
	testUnitGrpc = flag.Bool("grpc", false, "run all integration tests")
	testUnitCli = flag.Bool("cli", false, "run all integration tests")
	testCliCommand = flag.Bool("run-cli", false, "run all integration tests")
	testEverything = flag.Bool("everything", false, "run all integration tests")
)

func TestMain(m *testing.M) {

    flag.Parse()

	if *testCliCommand {
		main()
		return
	}
	if *testEverything {
		*testIntegration = true
		*testUnitGrpc = true
		*testUnitCli = true
	}

	if *testIntegration {
		*testIntegConsul = true
		*testIntegStatsd = true
	}

    if *testIntegConsul {
		setupServer()
        setupConsul()
    }
    if *testIntegStatsd {
		setupServer()
        setupStatsd()
    }
    os.Exit(m.Run())
}

func TestCli(t *testing.T) {
    cmd := exec.Command(os.Args[0], "-run-cli")
    cmd.Env = append(os.Environ(), "TEST_MAIN=crasher")
    err := cmd.Run()
    if e, ok := err.(*exec.ExitError); ok && !e.Success() {
        return
    }
    t.Fatalf("process err %v, want exit status 1", err)
}

func setupStatsd() {
}
func setupConsul() {
}
func setupServer() {
}

func TestNothing(t *testing.T) {
	Convey("Given something", t, func() {
		So("foo", ShouldEqual, "foo")
	})
}

