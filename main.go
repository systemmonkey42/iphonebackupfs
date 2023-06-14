package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	_ "github.com/mattn/go-sqlite3"
)

var progName = filepath.Base(os.Args[0])

type Globals struct {
	db          DB
	Debug       bool
	AllDomains  bool
	ListDomains bool
	Domain      string
	Root        string
	FSRoot      NodeEntry
}

var global Globals = Globals{}

func usage() {
	fmt.Fprintf(os.Stderr, "%s: invalid parameters\n", progName)
}

type DB struct {
	*sql.DB
}

func init() {
	flag.BoolVar(&global.AllDomains, "A", false, "Show all backup file domains.")
	flag.BoolVar(&global.ListDomains, "L", false, "List all domains in backup.")
	flag.BoolVar(&global.Debug, "v", false, "Verbose logging.")
	flag.StringVar(&global.Domain, "d", "CameraRollDomain", "Select domain to mount.")
}

func getBackupDir() (root string) {
	if flag.NArg() >= 0 {
		root = flag.Arg(0)
	}
	if root == "" {
		root = os.Getenv("ROOT")
	}
	return
}

func getMountPoint() (mount string) {
	if flag.NArg() >= 1 {
		mount = flag.Arg(1)
	}

	if mount == "" {
		mount = os.Getenv("MOUNT")
	}
	return
}

func openDB() error {
	global.Root = getBackupDir()

	debug("Opening database in %s", global.Root)
	err := global.db.OpenDB(global.Root)
	if err != nil {
		log.Fatalf("%s: %v", global.Root, err)
	}
	return err
}

func debug(fmt string, args ...any) {
	if global.Debug {
		log.Printf(fmt, args...)
	}
}

func main() {
	var err error
	log.SetFlags(0)
	log.SetPrefix(progName + ": ")
	flag.Parse()

	if flag.NArg() > 2 {
		usage()
		os.Exit(2)
	}

	switch {
	case global.ListDomains:

		err = openDB()
		if err != nil {
			log.Fatalf("%s: %v", global.Root, err)
		}

		domains, err := global.db.GetDomains()
		if err != nil {
			log.Fatal(err)
		}

		for d := range domains {
			fmt.Printf("%s\n", domains[d])
		}

	default:

		mountpoint := getMountPoint()
		if mountpoint == "" {
			log.Fatalf("mount point required.")
		}
		debug("Using mountpoint: %s\n", mountpoint)

		err = openDB()
		if err != nil {
			log.Fatalf("%s: %v", global.Root, err)
		}
		debug("Database opened successfully")

		global.FSRoot, err = global.db.ReadListing()

		if err != nil {
			log.Fatalf("%s: %v\n", global.Root, err)
		}

		if err = mount(global.Root, mountpoint); err != nil {
			log.Fatal(err)
		}
		debug("Completed.")
	}
}

func HandleSignals(mountpoint string) {
	ch := make(chan os.Signal, 1)
	go func() {
		for range ch {
			fmt.Printf("\rCtrl-C detected. Unmounting %s\n", mountpoint)
			debug("Unmounting filesystem")
			fuse.Unmount(mountpoint)
			signal.Stop(ch)
			return
		}
	}()
	debug("Enabling signal handlers")
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
}

func mount(path, mountpoint string) (err error) {
	c, err := fuse.Mount(mountpoint,
		fuse.FSName("iphone"),
		fuse.Subtype("iphonebackupfs"),
		//fuse.AllowOther(),
		fuse.ReadOnly(),
	)
	if err != nil {
		return err
	}
	debug("FUSE iniitiaalized")
	defer c.Close()

	filesys := &FS{DB: &DB{}, File: path}

	HandleSignals(mountpoint)

	debug("Serving files")
	if err := fs.Serve(c, filesys); err != nil {
		return err
	}
	debug("File server exited")

	return nil
}
