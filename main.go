package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	_ "github.com/mattn/go-sqlite3"
)

var progName = filepath.Base(os.Args[0])

func usage() {
	fmt.Fprintf(os.Stderr, "%s: invalid parameters\n", progName)
}

type DB struct {
	*sql.DB
}

func main() {
	log.SetFlags(0)
	log.SetPrefix(progName + ": ")
	flag.Parse()
	if flag.NArg() != 2 {
		usage()
		os.Exit(2)
	}

	root = flag.Arg(0)
	if root == "" {
		root = os.Getenv("ROOT")
	}
	mountpoint := flag.Arg(1)
	if err := mount(root, mountpoint); err != nil {
		log.Fatal(err)
	}
}

func mount(path, mountpoint string) (err error) {
	c, err := fuse.Mount(mountpoint,
		fuse.FSName("iphone"),
		fuse.Subtype("iphonebackupfs"),
	)
	if err != nil {
		return err
	}
	defer c.Close()

	filesys := &FS{DB: &DB{}, File: path}

	if err := fs.Serve(c, filesys); err != nil {
		return err
	}

	return nil
}

type FS struct {
	*DB
	File string
}

type DirNode struct {
	inode   uint64
	name    string
	entries map[string]NodeEntry
}

type FileNode struct {
	inode uint64
	name  string
	id    string
}

type FileHandle struct {
	fh io.ReadSeekCloser
}

var root string

var _ fs.FS = (*FS)(nil)
var _ fs.Node = (*DirNode)(nil)
var _ fs.Node = (*FileNode)(nil)
var _ fs.Handle = (*FileHandle)(nil)
var _ fs.HandleReleaser = (*FileHandle)(nil)

var _ = fs.NodeRequestLookuper(&DirNode{})
var _ = fs.NodeOpener(&FileNode{})
var _ = fs.HandleReadDirAller(&DirNode{})

type NodeEntry interface {
	fs.Node
	Add(uint64, string, string)
	Find(string) NodeEntry
	Name() string
	ID() string
	Dump()
	Inode() uint64
}

func (f *FileNode) Dump() {
	fmt.Printf(" %s [ %s ]\n", f.name, f.id)
}

func (f *FileNode) Fullname() string {
	file := filepath.Join(root, f.id[0:2], f.id)
	return file
}

func (f *FileNode) Inode() uint64 {
	return f.inode
}

func (d *DirNode) Inode() uint64 {
	return d.inode
}

func (d *DirNode) Dump() {
	fmt.Printf("%s /\n", d.name)
	for i := range d.entries {
		d.entries[i].Dump()
	}
}

func (f *FileNode) Add(inode uint64, id, path string) {
}

func (f *FileNode) Find(path string) NodeEntry {
	return f
}

func (f *FileNode) ID() string {
	return f.id
}

func (f *FileNode) Name() string {
	return f.name
}

func (d *DirNode) Add(inode uint64, id, path string) {
	p := strings.Split(path, "/")
	lp := len(p) - 1
	fp := d
	for i := range p {
		if i == lp {
			fp.entries[p[i]] = &FileNode{
				inode: inode,
				name:  p[i],
				id:    id,
			}
		} else {
			fn, ok := fp.entries[p[i]]
			if ok {
				fp = fn.(*DirNode)
			} else {
				fp.entries[p[i]] = &DirNode{
					inode:   inode,
					name:    p[i],
					entries: make(map[string]NodeEntry),
				}
				fp = fp.entries[p[i]].(*DirNode)
			}
		}
	}
	return
}

func (d *DirNode) Name() string {
	return d.name
}

func (d *DirNode) ID() string {
	return ""
}

func (d *DirNode) Find(path string) NodeEntry {
	return nil
}

func (d *DB) ReadListing() (NodeEntry, error) {

	r, err := d.Query("select fileid,relativepath,domain from files where flags=1")

	if err != nil {
		return nil, err
	}

	var dirs NodeEntry = &DirNode{
		entries: make(map[string]NodeEntry),
	}

	inode := uint64(1)
	for r.Next() {
		var id, path, domain string
		r.Scan(&id, &path, &domain)
		if domain == "CameraRollDomain" {
			dirs.Add(inode, id, path)
			inode++
		}
	}

	return dirs, nil
}

func (d *DB) OpenDB(file string) (list NodeEntry, err error) {
	d.DB, err = sql.Open("sqlite3", fmt.Sprintf("file:%s/Manifest.db?mode=ro", file))
	if err == nil {
		list, err = d.ReadListing()
	}
	return list, err
}

func (f *FS) Root() (n fs.Node, err error) {
	n = (*DirNode)(nil)

	if listing, err := f.DB.OpenDB(f.File); err == nil {
		return listing, nil
	}
	return nil, fuse.ENOENT
}

func (d *DirNode) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Atime = time.Now()
	attr.Ctime = attr.Atime
	attr.Mtime = attr.Atime
	attr.Size = uint64(len(d.entries))
	attr.Mode = os.ModeDir | 0755
	return nil
}

func (d *DirNode) Lookup(ctx context.Context, req *fuse.LookupRequest, resp *fuse.LookupResponse) (fs.Node, error) {
	if v, ok := d.entries[req.Name]; ok {
		return v, nil
	}
	return nil, fuse.ENOENT
}

func (f *FileNode) Attr(ctx context.Context, attr *fuse.Attr) error {
	file := f.Fullname()

	if info, err := os.Stat(file); err == nil {
		stat := info.Sys().(*syscall.Stat_t)

		attr.Mtime = info.ModTime()
		attr.Atime = time.Unix(int64(stat.Atim.Sec), int64(stat.Atim.Nsec))
		attr.Ctime = time.Unix(int64(stat.Ctim.Sec), int64(stat.Ctim.Nsec))
		attr.Size = uint64(info.Size())
		attr.Mode = info.Mode()
	} else {
		return err
	}
	return nil
}

func (f *FileNode) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	file := f.Fullname()

	if !req.Flags.IsReadOnly() {
		return nil, fuse.Errno(syscall.EACCES)
	}

	fh, err := os.Open(file)
	if err == nil {
		resp.Flags |= fuse.OpenDirectIO
		return &FileHandle{fh}, nil
	}

	return nil, err
}

func (f *FileHandle) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	err := f.fh.Close()
	return err
}

func (f *FileHandle) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	buf := make([]byte, req.Size)
	_, err := f.fh.Seek(req.Offset, os.SEEK_SET)
	if err != nil {
		return err
	}
	n, err := f.fh.Read(buf)
	if err != nil {
		return err
	}
	resp.Data = buf[:n]

	return err
}

func (d *DirNode) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {

	e := make([]fuse.Dirent, len(d.entries))

	f := 0
	for i := range d.entries {
		e[f].Inode = d.entries[i].Inode()
		e[f].Name = i
		switch d.entries[i].(type) {
		case *DirNode:
			e[f].Type = fuse.DT_Dir
		default:
			e[f].Type = fuse.DT_File
		}
		f++
	}

	return e, nil
}
