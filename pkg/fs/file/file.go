package file

import (
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type File struct {
	fs.Inode
	Attr fuse.Attr
	root *Root
}
