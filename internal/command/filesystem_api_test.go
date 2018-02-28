package command

import (
	"flag"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
	"syscall"
	"testing"

	"github.com/urfave/cli"

	"bitbucket.org/udt/wizefs/internal/config"
	"bitbucket.org/udt/wizefs/internal/globals"
)

const (
	packagePath = "internal/command"
)

var (
	projectPath = getProjectPath()
)

func getProjectPath() string {
	_, testFilename, _, _ := runtime.Caller(0)
	idx := strings.Index(testFilename, packagePath)
	return testFilename[0:idx]
}

func runCommand(t *testing.T, command string, origin string) (cerr error) {
	appPath := projectPath + "wizefs"
	c := exec.Command(appPath, command, origin)
	t.Logf("starting command %s...", command)
	cerr = c.Start()
	if cerr != nil {
		t.Errorf("starting command failed: %v", cerr)
	}

	t.Logf("waiting command %s...", command)
	cerr = c.Wait()

	t.Logf("finishing command %s...", command)
	return cerr
}

func assertExitCode(t *testing.T, cerr error, exitCode int) {
	if cerr != nil {
		if exiterr, ok := cerr.(*exec.ExitError); ok {
			if waitstat, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				if waitstat.ExitStatus() == exitCode {
					t.Logf("GREEN: Got expected exit status: %d.", exitCode)
				} else {
					t.Errorf("RED: Expected exit status: %d, Got: %d", exitCode, waitstat.ExitStatus())
				}
			}
		} else {
			t.Errorf("RED: Wait returned an unknown error: %v", cerr)
		}
	}
}

func TestCreateNormal(t *testing.T) {
	assertExitCode(t,
		runCommand(t, "create", "UNITTEST"),
		0)
}

func TestCreateInvalidOrigin(t *testing.T) {
	// TODO: create invalid origin before

	assertExitCode(t,
		runCommand(t, "create", "image.jpg"),
		globals.ExitOrigin)

	// TODO: remove invalid origin after
}

func TestCreateAlreadyExist(t *testing.T) {
	// create directory before
	os.MkdirAll(globals.OriginDirPath+"EXISTDIR", 0755)

	assertExitCode(t,
		runCommand(t, "create", "EXISTDIR"),
		globals.ExitOrigin)

	// remove directory after
	os.RemoveAll(globals.OriginDirPath + "EXISTDIR")
}

func TestDeleteNormal(t *testing.T) {
	assertExitCode(t,
		runCommand(t, "delete", "UNITTEST"),
		0)
}

//

// another approach for testing
func createMockContext(t *testing.T, mockArgs []string) *cli.Context {
	t.Log("Create mock context")
	mockApp := cli.NewApp()

	mockSet := flag.NewFlagSet("mock", 0)
	//mockArgs := []string{"TESTDIR"}
	mockSet.Parse(mockArgs)

	return cli.NewContext(mockApp, mockSet, nil)
}

func expect(t *testing.T, a interface{}, b interface{}) {
	if !reflect.DeepEqual(b, a) {
		t.Errorf("RED: Expected %#v (type %v) - Got %#v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	} else {
		t.Logf("GREEN: Expected %#v (type %v)", b, reflect.TypeOf(b))
	}
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
