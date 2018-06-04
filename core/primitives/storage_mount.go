package core

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
	"github.com/hanwen/go-fuse/zipfs"

	"bitbucket.org/udt/wizefs/core/fusefrontend"
	"bitbucket.org/udt/wizefs/core/globals"
	"bitbucket.org/udt/wizefs/core/syscallcompat"
	"bitbucket.org/udt/wizefs/core/tlog"
)

// DoMount mounts an directory.
// Called from main.
func (s *Storage) doMount(fstype globals.FSType,
	origin, originPath, mountpoint, mountpointPath string,
	notifypid int) (exitCode int, err error) {

	// Initialize FUSE server
	srv, exitCode, err := s.initFuseFrontend(fstype, originPath, mountpointPath)
	if exitCode != 0 || err != nil {

	}

	tlog.Debug.Println("Filesystem mounted and ready.")

	// TODO: do something with configuration
	if s.Config == nil {
		//config.InitWizeConfig()
		//} else {
		//	config.CommonConfig.Load()
	}
	err = s.Config.MountFilesystem(origin, mountpoint, mountpointPath)
	if err != nil {
		tlog.Warn.Printf("Problem with adding Filesystem to Config: %v", err)
	} else {
		err = s.Config.Save()
		if err != nil {
			tlog.Warn.Printf("Problem with saving Config: %v", err)
		}
	}

	tlog.Info.Printf("Filesystem added to configuration.")

	// FIXME: Mounting the Bucket
	s.buckets[origin].mounted = true
	s.buckets[origin].MountPoint = mountpoint

	tlog.Info.Printf("Bucket: %+v\n", s.buckets[origin])

	// We have been forked into the background, as evidenced by the set
	// "notifypid".
	if notifypid > 0 {
		// Chdir to the root directory so we don't block unmounting the CWD
		os.Chdir("/")

		// Daemons should redirect stdin, stdout and stderr
		s.redirectStdFds()

		// Disconnect from the controlling terminal by creating a new session.
		// This prevents us from getting SIGINT when the user presses Ctrl-C
		// to exit a running script that has called gocryptfs.
		_, err = syscall.Setsid()
		if err != nil {
			tlog.Warn.Printf("Setsid: %v", err)
		}
		// Send SIGUSR1 to our parent
		s.sendUsr1(notifypid)
	}

	// TODO: understand for what is it
	// Increase the open file limit to 4096. This is not essential, so do it after
	// we have switched to syslog and don't bother the user with warnings.
	s.setOpenFileLimit()

	// Wait for SIGINT in the background and unmount ourselves if we get it.
	// This prevents a dangling "Transport endpoint is not connected"
	// mountpoint if the user hits CTRL-C.
	s.handleSigint(srv, mountpointPath)

	// TODO: remove this?
	// Return memory that was allocated for scrypt (64M by default!) and other
	// stuff that is no longer needed to the OS
	debug.FreeOSMemory()

	// Jump into server loop. Returns when it gets an umount request from the kernel.
	srv.Serve()
	return 0, nil
}

// DoUnmount tries to umount "dir" and panics on error.
func (s *Storage) doUnmount(dir string) error {
	err := s.unmountErr(dir)
	if err != nil {
		//tlog.Warn.Println(err)
		//panic(err)
	}

	return err
}

// unmountErr tries to unmount "dir" and returns the resulting error.
func (s *Storage) unmountErr(dir string) error {
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
func (s *Storage) setOpenFileLimit() {
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
func (s *Storage) initFuseFrontend(fstype globals.FSType, originPath, mountpointPath string) (*fuse.Server, int, error) {
	// Reconciliate CLI and config file arguments into a fusefrontend.Args struct
	// that is passed to the filesystem implementation
	frontendArgs := fusefrontend.Args{
		OriginDir: originPath,
		Type:      fstype,
	}

	jsonBytes, _ := json.MarshalIndent(frontendArgs, "", "\t")
	tlog.Debug.Printf("frontendArgs: %s", string(jsonBytes))

	// Prepare root
	root := s.prepareRoot(frontendArgs)
	if root == nil {
		//os.Exit(globals.ExitType)
		return nil, globals.ExitType, nil
	}

	fuseOpts := &nodefs.Options{
		// These options are to be compatible with libfuse defaults,
		// making benchmarking easier.
		NegativeTimeout: time.Second,
		AttrTimeout:     time.Second,
		EntryTimeout:    time.Second,
		Debug:           true,
	}

	conn := nodefs.NewFileSystemConnector(root, fuseOpts)
	mountOpts := fuse.MountOptions{
		// Writes and reads are usually capped at 128kiB on Linux through
		// the FUSE_MAX_PAGES_PER_REQ kernel constant in fuse_i.h. Our
		// sync.Pool buffer pools are sized acc. to the default. Users may set
		// the kernel constant higher, and Synology NAS kernels are known to
		// have it >128kiB. We cannot handle more than 128kiB, so we tell
		// the kernel to limit the size explicitely.
		MaxWrite: fuse.MAX_KERNEL_WRITE,
		Options:  []string{fmt.Sprintf("max_read=%d", fuse.MAX_KERNEL_WRITE)},
		Debug:    fuseOpts.Debug,
	}

	// Set values shown in "df -T" and friends
	// First column, "Filesystem"
	mountOpts.FsName = tlog.ProgramName
	// Second column, "Type", will be shown as "fuse." + Name
	mountOpts.Name = tlog.ProgramName

	// Add a volume name if running osxfuse. Otherwise the Finder will show it as
	// something like "osxfuse Volume 0 (wizefs)".
	if runtime.GOOS == "darwin" {
		mountOpts.Options = append(mountOpts.Options, "volname="+path.Base(mountpointPath))
	}

	srv, err := fuse.NewServer(conn.RawFS(), mountpointPath, &mountOpts)
	if err != nil {
		tlog.Warn.Printf("fuse.NewServer failed: %v\n", err)
		if runtime.GOOS == "darwin" {
			tlog.Warn.Println("Maybe you should run: /Library/Filesystems/osxfuse.fs/Contents/Resources/load_osxfuse")
		}
		os.Exit(globals.ExitFuseNewServer)
		//return nil, globals.ExitFuseNewServer, err
	}

	// All FUSE file and directory create calls carry explicit permission
	// information. We need an unrestricted umask to create the files and
	// directories with the requested permissions.
	syscall.Umask(0000)

	return srv, 0, nil
}

// TODO: move to fusefrontend?
func (s *Storage) prepareRoot(args fusefrontend.Args) (root nodefs.Node) {
	switch args.Type {
	case globals.LoopbackFS, globals.LZFS:
		var finalFs pathfs.FileSystem

		// pathFsOpts are passed into go-fuse/pathfs
		pathFsOpts := &pathfs.PathNodeFsOptions{
			ClientInodes: true,
		}

		fs := fusefrontend.NewFS(args)
		finalFs = fs

		pathFs := pathfs.NewPathNodeFs(finalFs, pathFsOpts)

		root = pathFs.Root()

	case globals.ZipFS:
		// TODO: move to fusefrontend
		var err error
		root, err = zipfs.NewArchiveFileSystem(args.OriginDir)
		if err != nil {
			tlog.Warn.Printf("NewArchiveFileSystem failed: %v", err)
			os.Exit(globals.ExitOrigin)
		}

	default:
		tlog.Warn.Printf("Strange type of Filesystem: %d", args.Type)
		root = nil
	}

	return root
}

func (s *Storage) handleSigint(srv *fuse.Server, mountpoint string) {
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
		os.Exit(globals.ExitSigInt)
	}()
}

// redirectStdFds redirects stderr and stdout to syslog; stdin to /dev/null
func (s *Storage) redirectStdFds() {
	// Create a pipe pair "pw" -> "pr" and start logger reading from "pr".
	// We do it ourselves instead of using StdinPipe() because we need access
	// to the fd numbers.
	pr, pw, err := os.Pipe()
	if err != nil {
		tlog.Warn.Printf("redirectStdFds: could not create pipe: %v", err)
		return
	}
	tag := fmt.Sprintf("%s-%d-logger", tlog.ProgramName, os.Getpid())
	cmd := exec.Command("/usr/bin/logger", "-t", tag)
	cmd.Stdin = pr
	err = cmd.Start()
	if err != nil {
		tlog.Warn.Printf("redirectStdFds: could not start logger: %v", err)
		return
	}
	// The logger now reads on "pr". We can close it.
	pr.Close()
	// Redirect stout and stderr to "pw".
	err = syscallcompat.Dup3(int(pw.Fd()), 1, 0)
	if err != nil {
		tlog.Warn.Printf("redirectStdFds: stdout dup error: %v", err)
	}
	syscallcompat.Dup3(int(pw.Fd()), 2, 0)
	if err != nil {
		tlog.Warn.Printf("redirectStdFds: stderr dup error: %v", err)
	}
	// Our stout and stderr point to "pw". We can close the extra copy.
	pw.Close()
	// Redirect stdin to /dev/null
	nullFd, err := os.Open("/dev/null")
	if err != nil {
		tlog.Warn.Printf("redirectStdFds: could not open /dev/null: %v", err)
		return
	}
	err = syscallcompat.Dup3(int(nullFd.Fd()), 0, 0)
	if err != nil {
		tlog.Warn.Printf("redirectStdFds: stdin dup error: %v", err)
	}
	nullFd.Close()
}

// Send signal USR1 to "pid" (usually our parent process). This notifies it
// that the mounting has completed successfully.
func (s *Storage) sendUsr1(pid int) {
	p, err := os.FindProcess(pid)
	if err != nil {
		tlog.Warn.Printf("sendUsr1: FindProcess: %v", err)
		return
	}
	err = p.Signal(syscall.SIGUSR1)
	if err != nil {
		tlog.Warn.Printf("sendUsr1: Signal: %v", err)
	}
}
