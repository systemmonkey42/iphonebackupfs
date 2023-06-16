//go:build winfsp
// +build winfsp

package main

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/winfsp/cgofuse/fuse"
)

var (
	_ fuse.FileSystemInterface = (*FS)(nil)
	//_ fuse.FileSystemOpenEx    = (*FS)(nil)
	//_ fuse.FileSystemGetpath   = (*FS)(nil)
)

func mount(mountpoint string) (err error) {

	fs := NewFS()

	host := fuse.NewFileSystemHost(fs)

	// cgofuse handles Ctrl-C internally and unmounts the filesystem.
	// Just absorb it here.
	HandleSignals(func() {
	})

	host.Mount(mountpoint, nil)

	return nil
}

func split(path string) []string {
	return strings.Split(path, "/")
}

type FS struct {
	sync.Mutex
	*fuse.FileSystemBase

	root *FSNode
	open map[uint64]*FSNode
}

type FSNode struct {
	NodeEntry
	stat    fuse.Stat_t
	fh      *os.File
	opencnt int
}

func newFSFileNode(e NodeEntry, uid, gid uint32) *FSNode {

	file := e.Fullname()

	if info, err := os.Stat(file); err == nil {

		Size := int64(info.Size())
		Blocks := int64((Size + 511) / 512)

		t := fuse.NewTimespec(info.ModTime())

		return &FSNode{
			NodeEntry: e,
			stat: fuse.Stat_t{
				Dev:      0,
				Ino:      e.Inode(),
				Mode:     0640 | fuse.S_IFREG,
				Nlink:    1,
				Uid:      uid,
				Gid:      gid,
				Rdev:     0,
				Size:     Size,
				Atim:     t,
				Mtim:     t,
				Ctim:     t,
				Blksize:  512,
				Blocks:   Blocks,
				Birthtim: t,
				Flags:    0,
			},
		}
	}

	return nil
}

func newFSDirNode(e NodeEntry, uid, gid uint32) *FSNode {
	t := fuse.Now()

	return &FSNode{
		NodeEntry: e,

		stat: fuse.Stat_t{
			Ino:      e.Inode(),
			Mode:     0750 | fuse.S_IFDIR,
			Nlink:    2,
			Uid:      uid,
			Gid:      gid,
			Atim:     t,
			Mtim:     t,
			Ctim:     t,
			Birthtim: t,
			Flags:    0,
		},
	}
}

func newFSNode(e NodeEntry, uid, gid uint32) *FSNode {

	switch e.(type) {
	case *FileNode:
		return newFSFileNode(e, uid, gid)
	case *DirNode:
		return newFSDirNode(e, uid, gid)
	}
	return nil
}

func (fs *FS) makeNode(e NodeEntry) *FSNode {
	uid, gid, _ := fuse.Getcontext()
	return newFSNode(e, uid, gid)
}

func (fs *FS) lookupNode(path string) *FSNode {
	debug("FS:LookupNode Called: %s", path)
	e := fs.root.NodeEntry

	for _, c := range split(path) {
		if c != "" {
			switch e.(type) {
			case *DirNode:
				e = e.(*DirNode).entries[c]
			default:
				//fmt.Printf("-- LOOKUP NODE ABORT: %s\n", path)
				return nil
			}
		}
	}
	if e == nil {
		//fmt.Printf("Lookup returning node-not-found: %s\n", path)
		return nil
	}

	if n, ok := fs.open[e.Inode()]; ok {
		//fmt.Printf("Lookup returning existing node... \n")
		return n
	}

	//fmt.Printf("Lookup returning %#v\n", e.Name())
	return fs.makeNode(e)
}

func (fs *FS) getNode(path string, fh uint64) *FSNode {
	debug("FS:getNode Called")
	if ^uint64(0) == fh {
		node := fs.lookupNode(path)
		return node
	}

	return fs.open[fh]

}

// Sync - call as "defer fs.Sync()()" to handle lock and unlock semantics
func (fs *FS) Sync() func() {
	fs.Lock()
	return func() {
		fs.Unlock()
	}
}

type Node struct {
}

func NewFS() *FS {
	fs := new(FS)
	return fs
}

func (fs *FS) Init() {
	debug("FS:Init Called")
	defer fs.Sync()()

	fs.root = fs.makeNode(global.FSRoot)
	fs.open = make(map[uint64]*FSNode)
}

func (*FS) Destroy() {
	debug("FS:Destroy Called")
}

func (*FS) Statfs(path string, stat *fuse.Statfs_t) int {
	debug("FS:Statfs Called")
	return -fuse.ENOSYS
}

func (*FS) Mknod(path string, mode uint32, dev uint64) int {
	debug("FS:Mknod Called")
	return -fuse.ENOSYS
}

func (*FS) Mkdir(path string, mode uint32) int {
	debug("FS:Mkdir Called")
	return -fuse.ENOSYS
}

func (*FS) Unlink(path string) int {
	debug("FS:Unlink Called")
	return -fuse.ENOSYS
}

func (*FS) Rmdir(path string) int {
	debug("FS:Rmdir Called")
	return -fuse.ENOSYS
}

func (*FS) Link(oldpath string, newpath string) int {
	debug("FS:Link Called")
	return -fuse.ENOSYS
}

func (*FS) Symlink(target string, newpath string) int {
	debug("FS:Symlink Called")
	return -fuse.ENOSYS
}

func (*FS) Readlink(path string) (int, string) {
	debug("FS:Readlink Called")
	return -fuse.ENOSYS, ""
}

func (*FS) Rename(oldpath string, newpath string) int {
	debug("FS:Rename Called")
	return -fuse.ENOSYS
}

func (*FS) Chmod(path string, mode uint32) int {
	debug("FS:Chmod Called")
	return -fuse.ENOSYS
}

func (*FS) Chown(path string, uid uint32, gid uint32) int {
	debug("FS:Chown Called")
	return -fuse.ENOSYS
}

func (*FS) Utimens(path string, tmsp []fuse.Timespec) int {
	debug("FS:Utimens Called")
	return -fuse.ENOSYS
}

func (*FS) Access(path string, mask uint32) int {
	debug("FS:Access Called")
	return -fuse.ENOSYS
}

func (*FS) Create(path string, flags int, mode uint32) (int, uint64) {
	debug("FS:Create Called")
	return -fuse.ENOSYS, ^uint64(0)
}

func (fs *FS) Open(path string, flags int) (int, uint64) {
	debug("FS:Open Called")
	defer fs.Sync()()

	return fs.openNode(path, false)
}

func (fs *FS) Getattr(path string, stat *fuse.Stat_t, fh uint64) int {
	debug("FS:Getattr Called [%s] [%d]", path, fh)
	defer fs.Sync()()

	node := fs.getNode(path, fh)

	if node == nil {
		debug("FS:Getattr lookup failed, returning not-found: %s\n", path)
		return -fuse.ENOENT
	}

	*stat = node.stat
	return 0
}

func (*FS) Truncate(path string, size int64, fh uint64) int {
	debug("FS:Truncate Called")
	return -fuse.ENOSYS
}

func (fs *FS) Read(path string, buff []byte, ofst int64, fh uint64) (n int) {
	debug("FS:Read Called")
	defer fs.Sync()()

	node := fs.getNode(path, fh)

	node.fh.Seek(ofst, os.SEEK_SET)
	n, _ = node.fh.Read(buff)
	return
}

func (*FS) Write(path string, buff []byte, ofst int64, fh uint64) int {
	debug("FS:Write Called")
	return -fuse.ENOSYS
}

func (*FS) Flush(path string, fh uint64) int {
	debug("FS:Flush Called")
	return -fuse.ENOSYS
}

func (fs *FS) Release(path string, fh uint64) int {
	debug("FS:Release Called")
	defer fs.Sync()()

	return fs.closeNode(fh)
}

func (*FS) Fsync(path string, datasync bool, fh uint64) int {
	debug("FS:Fsync Called")
	return -fuse.ENOSYS
}

/*
func (*FS) Lock(path string, cmd int, lock *Lock_t, fh uint64) int {
 debug("FS:Lock Called")
	return -fuse.ENOSYS
}
*/

func (fs *FS) openNode(path string, isdir bool) (errc int, fh uint64) {

	var err error
	node := fs.lookupNode(path)

	if nil == node {
		return -fuse.ENOENT, ^uint64(0)
	}

	if 0 == node.opencnt {
		if !isdir {
			fn := node.NodeEntry.(*FileNode).Fullname()
			//fmt.Printf("Calling OS.OPEN on %s\n", fn)
			node.fh, err = os.Open(fn)
			if err != nil {
				fmt.Printf("RETURNING TOTAL FAILURE\n")
				return -fuse.EIO, ^uint64(0)
			}
		}
		fs.open[node.stat.Ino] = node
		node.opencnt++
	}

	return 0, node.stat.Ino
}

func (fs *FS) closeNode(fh uint64) int {
	debug("FS:closeNode Called [%d]", fh)
	if node, ok := fs.open[fh]; ok {
		if 0 < node.opencnt {
			node.opencnt--
		}
		if 0 == node.opencnt {
			fs.open[fh].fh.Close()
			delete(fs.open, fh)
		}
		return 0
	}
	return -fuse.EBADF
}

func (fs *FS) Opendir(path string) (int, uint64) {
	debug("FS:Opendir Called")
	defer fs.Sync()()

	err, fh := fs.openNode(path, true)
	return err, fh
}

func (fs *FS) Readdir(path string,
	fill func(name string, stat *fuse.Stat_t, ofst int64) bool,
	ofst int64,
	fh uint64) int {
	debug("FS:Readdir Called [%s] [%d]", path, fh)
	defer fs.Sync()()

	node := fs.getNode(path, fh)
	e := node.NodeEntry.(*DirNode).entries
	for i := range e {
		s := new(fuse.Stat_t)
		switch e[i].(type) {
		case *DirNode:
			s.Mode = 0744 | fuse.S_IFDIR
		case *FileNode:
			s.Mode = 0644
		}

		if !fill(i, s, 0) {
			fmt.Printf("Readdir - aborting\n")
			break
		}
	}
	return 0
}

func (fs *FS) Releasedir(path string, fh uint64) int {
	debug("FS:Releasedir Called")
	defer fs.Sync()()

	return fs.closeNode(fh)
}

func (*FS) Fsyncdir(path string, datasync bool, fh uint64) int {
	debug("FS:Fsyncdir Called")
	return -fuse.ENOSYS
}

func (*FS) Setxattr(path string, name string, value []byte, flags int) int {
	debug("FS:Setxattr Called")
	return -fuse.ENOSYS
}

var xattr map[string]int = map[string]int{
	"user.iphone.file":   0,
	"user.iphone.id":     1,
	"user.iphone.domain": 2,
}

func (fs *FS) Getxattr(path string, name string) (int, []byte) {
	debug("FS:Getxattr Called: [%s] [%s]", path, name)
	if id, ok := xattr[name]; ok {
		node := fs.lookupNode(path)
		if node == nil {
			return -fuse.ENOENT, nil
		}
		switch id {
		case 0:
			return 0, []byte(node.Fullname())
		case 1:
			return 0, []byte(node.ID())
		case 2:
			return 0, []byte(node.Domain())
		}
	}
	return -fuse.ENOSYS, nil
}

func (*FS) Removexattr(path string, name string) int {
	debug("FS:Removexattr Called")
	return -fuse.ENOSYS
}

func (fs *FS) Listxattr(path string, fill func(name string) bool) int {
	debug("FS:Listxattr Called [%s]", path)
	node := fs.lookupNode(path)
	if node == nil {
		debug("FS:Listxattr returning not-found")
		return -fuse.ENOENT
	}
	switch node.NodeEntry.(type) {
	case *DirNode:
		if !fill("user.iphone.domain") {
			return -fuse.ERANGE
		}
		return 0
	case *FileNode:
		for x := range xattr {
			if !fill(x) {
				return -fuse.ERANGE
			}
		}
	}
	return 0
}

var _ fuse.FileSystemInterface = (*FS)(nil)
