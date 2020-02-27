package file

import (
	"context"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/sirupsen/logrus"
)

const (
	dirSize = 4 * 1024 * 1024
)

type Root struct {
	fs.Inode
	Ino       uint64
	WorkSpace string
	mb        *MetaBlock

	sync.Mutex
}

func (r *Root) OpenDisk(filename string, mbytes int) (*os.File, error) {
	if _, err := os.Stat(filename); err == nil {
		if f, err := os.OpenFile(filename, os.O_RDWR, 0666); err == nil {
			return f, nil
		}
	}

	size := uint64(mbytes * 1024 * 1024)
	fd, _ := os.Create(filename)
	fd.Seek(int64(size-1), 0)
	fd.Write([]byte{0})
	fd.Seek(0, 0)
	r.mb = NewBlocks(fd, size, uint64(BLKSIZE))

	logrus.Infof("OpenDisk and create new disk: %s", filename)

	return fd, nil
}

func (r *Root) OnAdd(ctx context.Context) {
	value := uint64(time.Now().Unix())
	r.Lock()
	defer r.Unlock()

	r.Ino++
	cur := r.NewPersistentInode(
		ctx, &File{
			Attr: fuse.Attr{
				Size:  dirSize,
				Mode:  0644,
				Atime: value,
				Mtime: value,
				Ctime: value,
			},
		}, fs.StableAttr{Ino: r.Ino, Mode: syscall.S_IFDIR})

	r.AddChild(".", cur, false)

	r.Ino++
	pre := r.NewPersistentInode(
		ctx, &File{
			Attr: fuse.Attr{
				Size:  dirSize,
				Mode:  0644,
				Atime: value,
				Mtime: value,
				Ctime: value,
			},
		}, fs.StableAttr{Ino: r.Ino, Mode: syscall.S_IFDIR})
	r.AddChild("..", pre, false)
}

func (r *Root) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = 0755

	return 0
}

func (r *Root) Mkdir(ctx context.Context, name string, mode uint32, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	logrus.Infof("Mkdir name: %v, mode: %v, r: %v ", name, mode, r.String())

	r.Lock()
	defer r.Unlock()

	value := uint64(time.Now().Unix())

	r.Ino++
	ch := r.NewPersistentInode(
		ctx, &File{
			Attr: fuse.Attr{
				Size:  dirSize,
				Mode:  mode,
				Atime: value,
				Mtime: value,
				Ctime: value,
			},
		}, fs.StableAttr{Ino: r.Ino, Mode: syscall.S_IFDIR})
	r.AddChild(name, ch, false)

	f, _ := r.OpenDisk("/tmp/file.img", DISKSIZE)
	defer f.Close()
	r.mb.MetaToDisk(f)
	r.mb.WriteRootToDisk(f, r)

	logrus.Infof("zzlin Mkdir r: %#v", r)
	logrus.Infof("zzlin Mkdir mb: %#v", r.mb)
	logrus.Infof("zzlin Mkdir workspace: %#v", r.WorkSpace)
	logrus.Infof("zzlin Mkdir inode: %#v", r.Ino)
	logrus.Infof("zzlin Mkdir children: %#v", r.Children())

	return ch, 0
}

func (r *Root) Create(ctx context.Context, name string, flags uint32, mode uint32, out *fuse.EntryOut) (node *fs.Inode,
	fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {

	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	r.Ino++
	logrus.Infof("Root Create: %v, mode: %v, r: %v ", name, mode, r.String())

	value := uint64(time.Now().Unix())
	ch := r.NewPersistentInode(
		ctx, &File{
			Attr: fuse.Attr{
				Mode:  mode,
				Atime: value,
				Mtime: value,
				Ctime: value,
			},
		}, fs.StableAttr{
			Ino: r.Ino,
		})
	r.AddChild(name, ch, false)

	f, _ := r.OpenDisk("/tmp/file.img", DISKSIZE)
	defer f.Close()
	r.mb.MetaToDisk(f)
	r.mb.WriteRootToDisk(f, r)

	return ch, nil, 0, 0
}
