package util

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"runtime"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"

	"bitbucket.org/udt/wizefs/internal/config"
	"bitbucket.org/udt/wizefs/internal/exitcodes"
	"bitbucket.org/udt/wizefs/internal/fusefrontend"
	"bitbucket.org/udt/wizefs/internal/tlog"
)

// DoMount mounts an directory.
// Called from main.
func DoMount(origindir, mountpoint string, notifypid int) int {
	var err error

	// Initialize FUSE server
	srv := initFuseFrontend(origindir, mountpoint)
	tlog.Debug.Println("Filesystem mounted and ready.")

	// We have been forked into the background, as evidenced by the set
	// "notifypid".
	if notifypid > 0 {
		// Chdir to the root directory so we don't block unmounting the CWD
		os.Chdir("/")

		// Daemons should redirect stdin, stdout and stderr
		redirectStdFds()

		// Disconnect from the controlling terminal by creating a new session.
		// This prevents us from getting SIGINT when the user presses Ctrl-C
		// to exit a running script that has called gocryptfs.
		_, err = syscall.Setsid()
		if err != nil {
			tlog.Warn.Printf("Setsid: %v", err)
		}
		// Send SIGUSR1 to our parent
		sendUsr1(notifypid)
	}

	// Increase the open file limit to 4096. This is not essential, so do it after
	// we have switched to syslog and don't bother the user with warnings.
	setOpenFileLimit()

	// Wait for SIGINT in the background and unmount ourselves if we get it.
	// This prevents a dangling "Transport endpoint is not connected"
	// mountpoint if the user hits CTRL-C.
	handleSigint(srv, mountpoint)

	// Return memory that was allocated for scrypt (64M by default!) and other
	// stuff that is no longer needed to the OS
	debug.FreeOSMemory()

	// TODO: do something with configuration
	err = config.CommonConfig.AddFilesystem(origindir, mountpoint, 1)
	if err != nil {
		tlog.Warn.Printf("Problem with adding Filesystem to Config: %v", err)
	} else {
		err = config.CommonConfig.Save()
		if err != nil {
			tlog.Warn.Printf("Problem with saving Config: %v", err)
		}
	}

	tlog.Debug.Printf("Filesystem added to configuration.")

	// Jump into server loop. Returns when it gets an umount request from the kernel.
	srv.Serve()
	return 0
}

// DoUnmount tries to umount "dir" and panics on error.
func DoUnmount(dir string) {
	err := unmountErr(dir)
	if err != nil {
		tlog.Warn.Println(err)
		panic(err)
	}
}

// unmountErr tries to unmount "dir" and returns the resulting error.
func unmountErr(dir string) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "linux" {
		cmd = exec.Command("fusermount", "-u", dir)
	} else if runtime.GOOS == "darwin" {
		cmd = exec.Command("umount", dir)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// setOpenFileLimit tries to increase the open file limit to 4096 (the default hard
// limit on Linux).
func setOpenFileLimit() {
	var lim syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &lim)
	if err != nil {
		tlog.Warn.Printf("Getting RLIMIT_NOFILE failed: %v", err)
		return
	}
	if lim.Cur >= 4096 {
		return
	}
	lim.Cur = 4096
	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &lim)
	if err != nil {
		tlog.Warn.Printf("Setting RLIMIT_NOFILE to %+v failed: %v", lim, err)
		//         %+v output: "{Cur:4097 Max:4096}" ^
	}
}

// initFuseFrontend - initialize wizefs/fusefrontend
// Calls os.Exit on errors
func initFuseFrontend(origindir, mountpoint string) *fuse.Server {
	// Reconciliate CLI and config file arguments into a fusefrontend.Args struct
	// that is passed to the filesystem implementation
	frontendArgs := fusefrontend.Args{
		OriginDir: origindir,
	}

	jsonBytes, _ := json.MarshalIndent(frontendArgs, "", "\t")
	tlog.Debug.Printf("frontendArgs: %s", string(jsonBytes))
	var finalFs pathfs.FileSystem
	// pathFsOpts are passed into go-fuse/pathfs
	pathFsOpts := &pathfs.PathNodeFsOptions{ClientInodes: true}

	fs := fusefrontend.NewFS(frontendArgs)
	finalFs = fs

	pathFs := pathfs.NewPathNodeFs(finalFs, pathFsOpts)
	var fuseOpts *nodefs.Options

	fuseOpts = &nodefs.Options{
		// These options are to be compatible with libfuse defaults,
		// making benchmarking easier.
		NegativeTimeout: time.Second,
		AttrTimeout:     time.Second,
		EntryTimeout:    time.Second,
	}

	conn := nodefs.NewFileSystemConnector(pathFs.Root(), fuseOpts)
	mOpts := fuse.MountOptions{
		// Writes and reads are usually capped at 128kiB on Linux through
		// the FUSE_MAX_PAGES_PER_REQ kernel constant in fuse_i.h. Our
		// sync.Pool buffer pools are sized acc. to the default. Users may set
		// the kernel constant higher, and Synology NAS kernels are known to
		// have it >128kiB. We cannot handle more than 128kiB, so we tell
		// the kernel to limit the size explicitely.
		MaxWrite: fuse.MAX_KERNEL_WRITE,
		Options:  []string{fmt.Sprintf("max_read=%d", fuse.MAX_KERNEL_WRITE)},
	}

	// Set values shown in "df -T" and friends
	// First column, "Filesystem"
	//fsname := args.cipherdir
	//mOpts.Options = append(mOpts.Options, "fsname="+fsname)
	mOpts.FsName = tlog.ProgramName
	// Second column, "Type", will be shown as "fuse." + Name
	mOpts.Name = tlog.ProgramName

	// Add a volume name if running osxfuse. Otherwise the Finder will show it as
	// something like "osxfuse Volume 0 (wizefs)".
	if runtime.GOOS == "darwin" {
		mOpts.Options = append(mOpts.Options, "volname="+path.Base(mountpoint))
	}

	srv, err := fuse.NewServer(conn.RawFS(), mountpoint, &mOpts)
	if err != nil {
		tlog.Warn.Printf("fuse.NewServer failed: %v\n", err)
		if runtime.GOOS == "darwin" {
			tlog.Warn.Println("Maybe you should run: /Library/Filesystems/osxfuse.fs/Contents/Resources/load_osxfuse")
		}
		os.Exit(exitcodes.FuseNewServer)
	}
	srv.SetDebug(true)

	// All FUSE file and directory create calls carry explicit permission
	// information. We need an unrestricted umask to create the files and
	// directories with the requested permissions.
	syscall.Umask(0000)

	return srv
}

func handleSigint(srv *fuse.Server, mountpoint string) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	signal.Notify(ch, syscall.SIGTERM)
	go func() {
		<-ch
		err := srv.Unmount()
		if err != nil {
			tlog.Warn.Print(err)
			if runtime.GOOS == "linux" {
				// MacOSX does not support lazy unmount
				tlog.Warn.Println("Trying lazy unmount")
				cmd := exec.Command("fusermount", "-u", "-z", mountpoint)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Run()
			}
		}
		os.Exit(exitcodes.SigInt)
	}()
}
