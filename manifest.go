package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"unicode"

	_ "github.com/mattn/go-sqlite3"
)

var (
	uniqueID uint64
)

func nextID() uint64 {
	for {
		val := atomic.LoadUint64(&uniqueID)
		if atomic.CompareAndSwapUint64(&uniqueID, val, val+1) {
			return val
		}
	}
}

type NodeEntry interface {
	Add(string, string, string)
	Find(string) NodeEntry
	Fullname() string
	Name() string
	ID() string
	Domain() string
	Dump()
	Inode() uint64
}

type DirNode struct {
	inode   uint64
	name    string
	domain  string
	entries map[string]NodeEntry
}

type FileNode struct {
	inode  uint64
	name   string
	domain string
	id     string
}

type FileHandle struct {
	sync.Mutex
	fh    io.ReadSeekCloser
	inode uint64
	id    string
	pos   int64
}

func (f *FileHandle) Sync() func() {
	f.Lock()
	return func() {
		f.Unlock()
	}
}

func (f *FileNode) Dump() {
	debug("FileNode:Dump Called")
	fmt.Printf(" %s [ %s ]\n", f.name, f.id)
}

func (f *FileNode) Fullname() string {
	file := filepath.Join(global.Root, f.id[0:2], f.id)
	debug("FileNode:Fullname Called: %s\n", file)
	return file
}

func (f *FileNode) Inode() uint64 {
	debug("FileNode:Inode Called")
	return f.inode
}

func (d *DirNode) Inode() uint64 {
	debug("DirNode:Inode Called")
	return d.inode
}

func (d *DirNode) Fullname() string {
	debug("DirNode:Fullname Called")
	return ""
}

func (d *DirNode) Dump() {
	debug("DirNode:Dump Called")
	fmt.Printf("%s /\n", d.name)
	for i := range d.entries {
		d.entries[i].Dump()
	}
}

func (f *FileNode) Add(id, domain, path string) {
	debug("FileNode:Add Called")
}

func (f *FileNode) Find(path string) NodeEntry {
	debug("FileNode:Find Called")
	return f
}

func (f *FileNode) Domain() string {
	debug("FileNode:Domain Called")
	return f.domain
}

func (f *FileNode) ID() string {
	debug("FileNode:ID Called")
	return f.id
}

func (f *FileNode) Name() string {
	debug("FileNode:Name Called")
	return f.name
}

// Ugly function to convert "CameraRollDomain" into "Camera Roll", and AppDomain-com.vendor.games to "App/com.vendor.games"
func cleanDomain(d string) []string {

	w := false
	s := ""
	p := []string{}

	for _, l := range d {
		switch {
		case l == '-':
			if s != "" {

				if strings.HasSuffix(s, " Domain") {
					s = s[0 : len(s)-7]
				}

				p = append(p, s)
				s = ""
				w = false
			}

		default:
			if unicode.IsUpper(l) {
				if w {
					s = s + " "
				}
				w = false
			} else if unicode.IsLower(l) {
				w = true
			}
			s = s + string(l)
		}
	}
	if s != "" {

		if strings.HasSuffix(s, " Domain") {
			s = s[0 : len(s)-7]
		}

		p = append(p, s)
	}
	debug("Cleaned %s: %#v", d, p)
	return p
}

func (d *DirNode) Add(id, domain, path string) {
	debug("DirNode:Add Called: %s %-32s %s", id, domain, path)
	p := strings.Split(path, "/")
	fp := d

	// Handle "AllDomains" option by pre-pending domain name (after cleaning)
	if global.AllDomains {
		d := cleanDomain(domain)
		p = append(d, p...)
	}

	lp := len(p) - 1

	for i := range p {
		if i == lp {
			fp.entries[p[i]] = &FileNode{
				inode:  nextID(),
				name:   p[i],
				domain: domain,
				id:     id,
			}
		} else {
			fn, ok := fp.entries[p[i]]
			if ok {
				var ok bool
				// This error suggests a problem with "cleanDomain()" above resulting in duplicates
				if fp, ok = fn.(*DirNode); !ok {
					log.Fatalf("Found existing file where directory expected: %s", fn.(*FileNode).Fullname())
				}
			} else {
				fp.entries[p[i]] = &DirNode{
					inode:   nextID(),
					name:    p[i],
					domain:  domain,
					entries: make(map[string]NodeEntry),
				}
				fp = fp.entries[p[i]].(*DirNode)
			}
		}
	}
	return
}

func (d *DirNode) Domain() string {
	debug("DirNode:Domain Called")
	return d.domain
}

func (d *DirNode) Name() string {
	debug("DirNode:Name Called")
	return d.name
}

func (d *DirNode) ID() string {
	debug("DirNode:ID Called")
	return ""
}

func (d *DirNode) Find(path string) NodeEntry {
	debug("DirNode:Find Called")
	return nil
}

func (d *DB) GetDomains() ([]string, error) {
	debug("DB:GetDomains Called")
	r, err := d.Query("select distinct domain from files where flags=1")

	if err != nil {
		return nil, err
	}

	list := make([]string, 0, 100)
	for r.Next() {
		var domain string
		r.Scan(&domain)
		list = append(list, domain)
	}
	return list, nil
}

func (d *DB) ReadListing() (NodeEntry, error) {
	debug("DB:ReadListing Called")

	r, err := d.Query("select fileid,relativepath,domain from files where flags=1")

	if err != nil {
		return nil, err
	}

	var dirs NodeEntry = &DirNode{
		entries: make(map[string]NodeEntry),
	}

	for r.Next() {
		var id, path, domain string
		r.Scan(&id, &path, &domain)
		if global.AllDomains || domain == global.Domain {
			dirs.Add(id, domain, path)
		}
	}

	return dirs, nil
}

func (d *DB) OpenDB(file string) (err error) {
	debug("DB:OpenDB Called")
	d.DB, err = sql.Open("sqlite3", fmt.Sprintf("file:%s/Manifest.db?immutable=1&mode=ro", file))

	if err != nil {
		panic(err)
	}

	return err
}
