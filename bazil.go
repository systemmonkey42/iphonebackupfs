//go:build bazil
// +build bazil

package main

import (
	"context"
	"io"
	"os"
	"syscall"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

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

	filesys := &FS{}

	HandleSignals(mountpoint)

	debug("Serving files")
	if err := fs.Serve(c, filesys); err != nil {
		return err
	}
	debug("File server exited")

	return nil
}

func unmount(mountpoint string) (err error) {
	err = fuse.Unmount(mountpoint)
	return
}

type FS struct {
}

type FSDir struct {
	NodeEntry
}

type FSFile struct {
	NodeEntry
}

var _ fs.FS = (*FS)(nil)
var _ fs.Node = (*FSDir)(nil)
var _ fs.Node = (*FSFile)(nil)
var _ fs.Handle = (*FileHandle)(nil)
var _ fs.HandleReleaser = (*FileHandle)(nil)
var _ = fs.NodeRequestLookuper(&FSDir{})
var _ = fs.NodeOpener(&FSFile{})
var _ = fs.HandleReadDirAller(&FSDir{})

func (f *FS) Root() (n fs.Node, err error) {
	debug("FS:Root Called")
	root := &FSDir{global.FSRoot}
	return root, nil
}

func (f *FSDir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	debug("DirNode:ReadDirAll Called")

	e := f.NodeEntry.(*DirNode).entries
	r := make([]fuse.Dirent, len(e))

	ri := 0
	for i := range e {
		r[ri].Inode = e[i].Inode()
		r[ri].Name = i
		switch e[i].(type) {
		case *DirNode:
			r[ri].Type = fuse.DT_Dir
		default:
			r[ri].Type = fuse.DT_File
		}
		ri++
	}

	return r, nil
}

func (f *FSDir) Lookup(ctx context.Context, req *fuse.LookupRequest, resp *fuse.LookupResponse) (fs.Node, error) {
	debug("DirNode:Lookup Called")
	if v, ok := f.NodeEntry.(*DirNode).entries[req.Name]; ok {
		switch v.(type) {
		case *FileNode:
			return &FSFile{v}, nil
		case *DirNode:
			return &FSDir{v}, nil
		}
	}
	return nil, fuse.ENOENT
}

func (f *FSFile) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	debug("FileNode:Open Called")
	e := f.NodeEntry.(*FileNode)
	file := e.Fullname()

	if !req.Flags.IsReadOnly() {
		return nil, fuse.Errno(syscall.EACCES)
	}

	fh, err := os.Open(file)
	if err == nil {
		//resp.Flags |= fuse.OpenDirectIO
		return &FileHandle{fh: fh, inode: e.inode, id: e.id}, nil
	}

	return nil, err
}

func (f *FSFile) Attr(ctx context.Context, attr *fuse.Attr) error {
	debug("FileNode:Attr Called")
	file := f.NodeEntry.(*FileNode).Fullname()

	if info, err := os.Stat(file); err == nil {
		stat := info.Sys().(*syscall.Stat_t)

		attr.Mtime = info.ModTime()
		attr.Atime = time.Unix(int64(stat.Atim.Sec), int64(stat.Atim.Nsec))
		attr.Ctime = time.Unix(int64(stat.Ctim.Sec), int64(stat.Ctim.Nsec))
		attr.Size = uint64(info.Size())
		attr.Mode = info.Mode()
	} else {
		debug("Stat(%s) error: %v", file, err)
		return err
	}
	return nil
}

func (f *FSDir) Attr(ctx context.Context, attr *fuse.Attr) error {
	debug("DirNode:Attr Called")

	e := f.NodeEntry.(*DirNode).entries

	attr.Atime = time.Now()
	attr.Ctime = attr.Atime
	attr.Mtime = attr.Atime
	attr.Size = uint64(len(e))
	attr.Mode = os.ModeDir | 0755
	return nil
}

func (f *FileHandle) Release(ctx context.Context, req *fuse.ReleaseRequest) (err error) {
	debug("FileHandle:Release Called")
	err = f.fh.Close()
	return
}

func (f *FileHandle) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) (err error) {
	debug("FileHandle:Read Called")
	f.Lock()
	defer f.Unlock()

	_, err = f.fh.Seek(req.Offset, os.SEEK_SET)
	if err != nil {
		return
	}

	buf := make([]byte, req.Size)
	n, err := f.fh.Read(buf)
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}
	resp.Data = buf[:n]
	return err
}
