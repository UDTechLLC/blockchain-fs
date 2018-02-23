package command

import (
	"flag"
	"os/exec"
	"reflect"
	"syscall"
	"testing"

	"github.com/urfave/cli"

	"bitbucket.org/udt/wizefs/internal/config"
)

const (
	projectPath = "/home/sergey/code/go/src/bitbucket.org/udt/wizefs/"
)

func testCommand(t *testing.T, command string, origin string) {
	appPath := projectPath + "wizefs"
	c := exec.Command(appPath, command, origin)
	t.Logf("starting command %s...", command)
	cerr := c.Start()
	if cerr != nil {
		t.Errorf("starting command failed: %v", cerr)
	}

	t.Logf("waiting command %s...", command)
	cerr = c.Wait()
	if cerr != nil {
		if exiterr, ok := cerr.(*exec.ExitError); ok {
			if waitstat, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				t.Logf("wait returned an exit status: %d", waitstat.ExitStatus())
			}
		} else {
			t.Errorf("wait returned an unknown error: %v", cerr)
		}
	}
	t.Logf("finishing command %s...", command)
}

func TestCommandCreate(t *testing.T) {
	testCommand(t, "create", "UNITTEST")
}

func TestCommandDelete(t *testing.T) {
	testCommand(t, "delete", "UNITTEST")
}

// another approach for testing
func createMockContext(t *testing.T, mockArgs []string) *cli.Context {
	t.Log("Create mock context")
	mockApp := cli.NewApp()

	mockSet := flag.NewFlagSet("mock", 0)
	//mockArgs := []string{"TESTDIR"}
	mockSet.Parse(mockArgs)

	return cli.NewContext(mockApp, mockSet, nil)
}

func testCmdCreateFilesystem(t *testing.T) {
	// INIT
	config.InitWizeConfig()

	// GIVEN
	mockCtx := createMockContext(t, []string{"UNITTEST"})

	// WHEN
	err := CmdCreateFilesystem(mockCtx)

	// SHOULD
	expect(t, err, nil)
}

// to test different use cases
func testCmdDeleteFilesystem(t *testing.T) {
	t.Log("TODO")
}

// to test different use cases
func testCmdMountFilesystem(t *testing.T) {
	t.Log("TODO")
}

// Now this test will not work because we should daemonize mount command
// So we can test it only with bash tests
func testCmdMountLZFS1(t *testing.T) {
	// INIT
	config.InitWizeConfig()

	// GIVEN
	// How to add supporting --fg and --notifypid to ForkChild?
	mockCtx := createMockContext(t, []string{"_test.zip"})

	// WHEN
	err := CmdMountFilesystem(mockCtx)
	t.Logf("Error: %v", err)

	// SHOULD - just two examples of expectations

	//errmessage := ""
	//if err != nil {
	//	errmessage = err.Error()
	//}
	//expect(t, errmessage, "LZFS files are not support now")

	//expect(t, err, nil)
}

// to test different use cases
func testCmdUnmountFilesystem(t *testing.T) {
	t.Log("TODO")
}

// to test different use cases
func testFullCycleFilesystem(t *testing.T) {
	t.Log("TODO")
}

func expect(t *testing.T, a interface{}, b interface{}) {
	if !reflect.DeepEqual(b, a) {
		t.Errorf("RED: Expected %#v (type %v) - Got %#v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	} else {
		t.Logf("GREEN: Expected %#v (type %v)", b, reflect.TypeOf(b))
	}
}
