//go:build winfsp
// +build winfsp

package main

import (
	"fmt"
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
	host.Mount("", []string{mountpoint})

	return nil
}

func unmount(mountpoint string) (err error) {
	//err = fuse.Unmount(mountpoint)
	return
}

func split(path string) []string {
	return strings.Split(path, "/")
}

type FS struct {
	sync.Mutex
	*fuse.FileSystemBase

	ino  uint64
	root *FSNode
	open map[uint64]*FSNode
}

type FSNode struct {
	NodeEntry
	stat    fuse.Stat_t
	opencnt int
}

func NewFSNode(e NodeEntry, ino uint64, uid, gid uint32) *FSNode {
	node := &FSNode{
		NodeEntry: e,
	}
	if filenode, ok := e.(*FileNode); ok {
		fmt.Printf("File = %s\n", filenode.Fullname())
	} else {
		fmt.Printf("Directory details\n")
		tmsp := fuse.Now()
		node.stat = fuse.Stat_t{
			Ino:      ino,
			Mode:     0644 | fuse.S_IFDIR,
			Nlink:    2,
			Uid:      uid,
			Gid:      gid,
			Atim:     tmsp,
			Mtim:     tmsp,
			Ctim:     tmsp,
			Birthtim: tmsp,
			Flags:    0,
		}
	}

	return node
}

func (fs *FS) makeNode(e NodeEntry) *FSNode {
	uid, gid, _ := fuse.Getcontext()
	return NewFSNode(global.FSRoot, fs.ino, uid, gid)
}

func (fs *FS) LookupNode(path string) *FSNode {
	debug("FS:LookupNode Called")
	n := fs.root

	fmt.Printf("Lookup path [%s]\n", path)
	for _, c := range split(path) {
		fmt.Printf("Checking for %s\n", c)
	}
	return n
}

func (fs *FS) getNode(path string, fh uint64) *FSNode {
	if ^uint64(0) == fh {
		node := fs.LookupNode(path)
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
	fs.ino = 0

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

func (*FS) Open(path string, flags int) (int, uint64) {
	debug("FS:Open Called")
	return -fuse.ENOSYS, ^uint64(0)
}

func (fs *FS) Getattr(path string, stat *fuse.Stat_t, fh uint64) int {
	debug("FS:Getattr Called")
	fmt.Printf("Getattr: path = %s\n", path)
	defer fs.Sync()()

	node := fs.getNode(path, fh)
	if node == nil {
		return -fuse.ENOENT
	}

	*stat = node.stat
	return 0
}

func (*FS) Truncate(path string, size int64, fh uint64) int {
	debug("FS:Truncate Called")
	return -fuse.ENOSYS
}

func (*FS) Read(path string, buff []byte, ofst int64, fh uint64) int {
	debug("FS:Read Called")
	return -fuse.ENOSYS
}

func (*FS) Write(path string, buff []byte, ofst int64, fh uint64) int {
	debug("FS:Write Called")
	return -fuse.ENOSYS
}

func (*FS) Flush(path string, fh uint64) int {
	debug("FS:Flush Called")
	return -fuse.ENOSYS
}

func (*FS) Release(path string, fh uint64) int {
	debug("FS:Release Called")
	return -fuse.ENOSYS
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

	node := fs.LookupNode(path)

	node.opencnt++

	if 1 == node.opencnt {
		fs.open[node.stat.Ino] = node
	}

	return 0, node.stat.Ino
	//return -fuse.ENOSYS, ^uint64(0)
}

func (fs *FS) closeNode(fh uint64) int {
	debug("FS:closeNode Called")
	if node, ok := fs.open[fh]; ok {
		if 0 < node.opencnt {
			node.opencnt--
		}
		if 0 == node.opencnt {
			delete(fs.open, fh)
		}
		return 0
	}
	fmt.Printf("closeNode: returning failz\n")
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
	debug("FS:Readdir Called")
	defer fs.Sync()()

	node := fs.getNode(path, fh)
	e := node.NodeEntry.(*DirNode).entries
	for i := range e {
		fmt.Printf("i = %v\n", i)
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

func (*FS) Getxattr(path string, name string) (int, []byte) {
	debug("FS:Getxattr Called")
	return -fuse.ENOSYS, nil
}

func (*FS) Removexattr(path string, name string) int {
	debug("FS:Removexattr Called")
	return -fuse.ENOSYS
}

func (*FS) Listxattr(path string, fill func(name string) bool) int {
	debug("FS:Listxattr Called")
	return -fuse.ENOSYS
}

var _ fuse.FileSystemInterface = (*FS)(nil)
