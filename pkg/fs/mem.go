package fs

import (
	"context"
	"log"
	"sync"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/sirupsen/logrus"
)

const (
	b  = 1024
	kb = 1024 * 1024

	dirKB = 4 * 1024 * 1024
)

type MemoryFileSystem struct {
	Debug     bool
	WorkSpace string
}

func NewMemoryFileSystem(workSpace string, debug bool) *MemoryFileSystem {
	return &MemoryFileSystem{
		Debug:     debug,
		WorkSpace: workSpace,
	}
}

func (m *MemoryFileSystem) Create() {
	root := &MemRoot{}

	logrus.Infof("MemoryFileSystem create workspace: %s begin", m.WorkSpace)

	opts := &fs.Options{}
	opts.Debug = m.Debug
	server, err := fs.Mount(m.WorkSpace, root, opts)
	if err != nil {
		log.Fatalf("Mount fail: %v\n", err)
	}

	go server.Wait()

	logrus.Infof("MemoryFileSystem create workspace: %s end", m.WorkSpace)
}

type MemRoot struct {
	sync.Mutex
	fs.Inode
}

// 需要枷锁
var Ino uint64 = 1

func (r *MemRoot) OnAdd(ctx context.Context) {
	Ino++
	ch := r.NewPersistentInode(
		ctx, &MemRegularFile{
			Attr: fuse.Attr{
				Mode: 0644,
			},
		}, fs.StableAttr{Ino: Ino})

	r.AddChild(".", ch, false)

	// writeToMetaData(r.file, r)

}

func (r *MemRoot) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = 0755

	return 0
}

func (r *MemRoot) Mkdir(ctx context.Context, name string, mode uint32, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	logrus.Infof("MemRoot Mkdir name: %v, mode: %v, r: %v ", name, mode, r.String())

	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	value := uint64(time.Now().Unix())

	Ino++
	ch := r.NewPersistentInode(
		ctx, &MemRegularFile{
			Attr: fuse.Attr{
				Size:  dirKB,
				Mode:  mode,
				Atime: value,
				Mtime: value,
				Ctime: value,
			},
		}, fs.StableAttr{Ino: Ino, Mode: syscall.S_IFDIR})
	r.AddChild(name, ch, false)

	return ch, 0
}

func (r *MemRoot) Create(ctx context.Context, name string, flags uint32, mode uint32, out *fuse.EntryOut) (node *fs.Inode,
	fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {

	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	Ino++
	logrus.Infof("MemRoot Create: %v, mode: %v, r: %v ", name, mode, r.String())

	value := uint64(time.Now().Unix())
	ch := r.NewPersistentInode(
		ctx, &MemRegularFile{
			Attr: fuse.Attr{
				Size:  dirKB,
				Mode:  mode,
				Atime: value,
				Mtime: value,
				Ctime: value,
			},
		}, fs.StableAttr{
			Ino: Ino,
		})
	r.AddChild(name, ch, false)

	return ch, nil, 0, 0
}

var _ = (fs.NodeGetattrer)((*MemRoot)(nil))
var _ = (fs.NodeOnAdder)((*MemRoot)(nil))

// MemRegularFile is a filesystem node that holds a read-only data
// slice in memory.
var I uint64 = 100000

type MemRegularFile struct {
	fs.Inode

	mu   sync.Mutex
	Data []byte
	Attr fuse.Attr
}

var _ = (fs.NodeOpener)((*MemRegularFile)(nil))
var _ = (fs.NodeReader)((*MemRegularFile)(nil))
var _ = (fs.NodeWriter)((*MemRegularFile)(nil))
var _ = (fs.NodeSetattrer)((*MemRegularFile)(nil))
var _ = (fs.NodeFlusher)((*MemRegularFile)(nil))

func (f *MemRegularFile) Open(ctx context.Context, flags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	return nil, fuse.FOPEN_KEEP_CACHE, fs.OK
}

func (f *MemRegularFile) Write(ctx context.Context, fh fs.FileHandle, data []byte, off int64) (uint32, syscall.Errno) {
	logrus.Infof("MemRegularFile: %s Write off: %v", f.String(), off)
	f.mu.Lock()
	defer f.mu.Unlock()
	end := int64(len(data)) + off
	if int64(len(f.Data)) < end {
		n := make([]byte, end)
		copy(n, f.Data)
		f.Data = n
	}

	copy(f.Data[off:off+int64(len(data))], data)

	return uint32(len(data)), 0
}

var _ = (fs.NodeGetattrer)((*MemRegularFile)(nil))

func (f *MemRegularFile) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	f.mu.Lock()
	defer f.mu.Unlock()
	out.Attr = f.Attr
	out.Attr.Size = uint64(len(f.Data))
	return fs.OK
}

func (f *MemRegularFile) Setattr(ctx context.Context, fh fs.FileHandle, in *fuse.SetAttrIn, out *fuse.AttrOut) syscall.Errno {
	f.mu.Lock()
	defer f.mu.Unlock()
	if sz, ok := in.GetSize(); ok {
		f.Data = f.Data[:sz]
	}
	out.Attr = f.Attr
	out.Size = uint64(len(f.Data))
	return fs.OK
}

func (f *MemRegularFile) Flush(ctx context.Context, fh fs.FileHandle) syscall.Errno {
	return 0
}

func (f *MemRegularFile) Read(ctx context.Context, fh fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	log.Printf("MemRegularFile: %s Read off: %v", f.String(), off)

	f.mu.Lock()
	defer f.mu.Unlock()
	end := int(off) + len(dest)
	if end > len(f.Data) {
		end = len(f.Data)
	}

	return fuse.ReadResultData(f.Data[off:end]), fs.OK
}

func (f *MemRegularFile) Create(ctx context.Context, name string, flags uint32, mode uint32, out *fuse.EntryOut) (node *fs.Inode,
	fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	logrus.Infof("MemRegularFile: Create name： %s", name)

	I++
	ch := f.NewPersistentInode(
		ctx, &MemRegularFile{
			Attr: fuse.Attr{
				Mode: mode,
			},
		}, fs.StableAttr{Ino: I})
	f.AddChild(name, ch, false)

	return ch, nil, 0, 0
}

func (f *MemRegularFile) Mkdir(ctx context.Context, name string, mode uint32, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	log.Printf("MemRegularFile Mkdir name: %v   stable: %v", name, f.Attr.Ino)

	I++
	ch := f.NewPersistentInode(
		ctx, &MemRegularFile{
			Attr: fuse.Attr{
				Mode: mode,
			},
		}, fs.StableAttr{Ino: I, Mode: syscall.S_IFDIR})
	f.AddChild(name, ch, false)

	return ch, 0
}

// MemSymlink is an inode holding a symlink in memory.
type MemSymlink struct {
	fs.Inode
	Attr fuse.Attr
	Data []byte
}

var _ = (fs.NodeReadlinker)((*MemSymlink)(nil))

func (l *MemSymlink) Readlink(ctx context.Context) ([]byte, syscall.Errno) {
	return l.Data, fs.OK
}

var _ = (fs.NodeGetattrer)((*MemSymlink)(nil))

func (l *MemSymlink) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Attr = l.Attr
	return fs.OK
}
