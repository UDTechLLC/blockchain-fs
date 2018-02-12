package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"

	"github.com/leedark/storage-system/internal/exitcodes"
	"github.com/leedark/storage-system/internal/fusefrontend"
)

// doMount mounts an directory.
// Called from main.
func doMount(args *argContainer) int {
	// Check mountpoint
	var err error
	args.mountpoint, err = filepath.Abs(flagSet.Arg(1))
	if err != nil {
		fmt.Println("Invalid mountpoint: %v", err)
		os.Exit(exitcodes.MountPoint)
	}

	// Initialize FUSE server
	srv := initFuseFrontend(args)
	fmt.Println("Filesystem mounted and ready.")

	// Wait for SIGINT in the background and unmount ourselves if we get it.
	// This prevents a dangling "Transport endpoint is not connected"
	// mountpoint if the user hits CTRL-C.
	handleSigint(srv, args.mountpoint)

	// Return memory that was allocated for scrypt (64M by default!) and other
	// stuff that is no longer needed to the OS
	debug.FreeOSMemory()

	// Jump into server loop. Returns when it gets an umount request from the kernel.
	srv.Serve()
	return 0
}

// initFuseFrontend - initialize storage-system/fusefrontend
// Calls os.Exit on errors
func initFuseFrontend(args *argContainer) *fuse.Server {
	// Reconciliate CLI and config file arguments into a fusefrontend.Args struct
	// that is passed to the filesystem implementation
	frontendArgs := fusefrontend.Args{
		Cipherdir: args.cipherdir,
	}

	jsonBytes, _ := json.MarshalIndent(frontendArgs, "", "\t")
	fmt.Printf("frontendArgs: %s\n", string(jsonBytes))
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
	mOpts.FsName = "storage-system"
	// Second column, "Type", will be shown as "fuse." + Name
	mOpts.Name = "storage-system"

	// Add a volume name if running osxfuse. Otherwise the Finder will show it as
	// something like "osxfuse Volume 0 (storage-system)".
	if runtime.GOOS == "darwin" {
		mOpts.Options = append(mOpts.Options, "volname="+path.Base(args.mountpoint))
	}

	srv, err := fuse.NewServer(conn.RawFS(), args.mountpoint, &mOpts)
	if err != nil {
		fmt.Println("fuse.NewServer failed: %v", err)
		if runtime.GOOS == "darwin" {
			fmt.Println("Maybe you should run: /Library/Filesystems/osxfuse.fs/Contents/Resources/load_osxfuse")
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
			fmt.Print(err)
			if runtime.GOOS == "linux" {
				// MacOSX does not support lazy unmount
				fmt.Println("Trying lazy unmount")
				cmd := exec.Command("fusermount", "-u", "-z", mountpoint)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Run()
			}
		}
		os.Exit(exitcodes.SigInt)
	}()
}
