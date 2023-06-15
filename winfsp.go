//go:build winfsp
// +build winfsp

package main

import (
	"github.com/winfsp/cgofuse/fuse"
)

var (
	_ fuse.FileSystemInterface = (*FS)(nil)
	//_ fuse.FileSystemOpenEx    = (*FS)(nil)
	//_ fuse.FileSystemGetpath   = (*FS)(nil)
)

//func mount(path, mountpoint string) (err error) {
//	c, err := fuse.Mount(mountpoint,
//		fuse.FSName("iphone"),
//		fuse.Subtype("iphonebackupfs"),
//		//fuse.AllowOther(),
//		fuse.ReadOnly(),
//	)
//	if err != nil {
//		return err
//	}
//	debug("FUSE iniitiaalized")
//	defer c.Close()
//
//	filesys := &FS{DB: &DB{}, File: path}
//
//	HandleSignals(mountpoint)
//
//	debug("Serving files")
//	if err := fs.Serve(c, filesys); err != nil {
//		return err
//	}
//	debug("File server exited")
//
//	return nil
//}

type FS struct {
	*fuse.FileSystemBase
	*DB
	File string
}
