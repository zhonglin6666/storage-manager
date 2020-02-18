package fs

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"sync"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/sirupsen/logrus"
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

	logrus.Debugf("MemoryFileSystem create workspace: %s begin", m.WorkSpace)

	opts := &fs.Options{}
	opts.Debug = m.Debug
	server, err := fs.Mount(m.WorkSpace, root, opts)
	if err != nil {
		log.Fatalf("Mount fail: %v\n", err)
	}

	server.Wait()

	logrus.Debugf("MemoryFileSystem create workspace: %s end", m.WorkSpace)
}

type MemRoot struct {
	file *os.File
	fs.Inode
}

// 需要枷锁
var Ino uint64 = 1

func (r *MemRoot) OnAdd(ctx context.Context) {
	Ino++
	ch := r.NewPersistentInode(
		ctx, &MemRegularFile{
			Data: []byte("aaaaaaaaaa\n"),
			Attr: fuse.Attr{
				Mode: 0644,
			},
		}, fs.StableAttr{Ino: Ino})

	r.AddChild("file.log", ch, false)

	writeToMetaData(r.file, r)

}

func writeToMetaData(f *os.File, r *MemRoot) error {
	data, err := json.Marshal(r)
	if err != nil {
		log.Printf("writeToMetaData Marshal error: %v", err)
		return err
	}

	log.Printf("writeToMetaData len: %v", len(data))

	_, err = f.WriteAt([]byte(data), int64(0))
	if err != nil {
		log.Printf("writeToMetaData WriteAt error: %v", err)
		return err
	}
	return nil
}

func (r *MemRoot) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = 0755

	return 0
}

func (r *MemRoot) Mkdir(ctx context.Context, name string, mode uint32, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	log.Printf("HelloRoot Mkdir name: %v, mode: %v, r: %v ", name, mode, r.String())

	Ino++
	ch := r.NewPersistentInode(
		ctx, &MemRegularFile{
			Attr: fuse.Attr{
				Mode: mode,
			},
		}, fs.StableAttr{Ino: Ino, Mode: syscall.S_IFDIR})
	r.AddChild(name, ch, false)

	return ch, 0
}

func (r *MemRoot) Create(ctx context.Context, name string, flags uint32, mode uint32, out *fuse.EntryOut) (node *fs.Inode,
	fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {

	Ino++
	ch := r.NewPersistentInode(
		ctx, &fs.MemRegularFile{
			Attr: fuse.Attr{
				Mode: mode,
			},
		}, fs.StableAttr{Ino: Ino})
	r.AddChild(name, ch, false)

	// file := fs.NewLoopbackFile(int(r.file.Fd()))

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
	f.mu.Lock()
	defer f.mu.Unlock()
	end := int64(len(data)) + off
	if int64(len(f.Data)) < end {
		n := make([]byte, end)
		copy(n, f.Data)
		f.Data = n
	}

	log.Printf("MemRegularFile Write len: %v off: %v", len(data), off)

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
	f.mu.Lock()
	defer f.mu.Unlock()
	end := int(off) + len(dest)
	if end > len(f.Data) {
		end = len(f.Data)
	}

	log.Printf("MemRegularFile Read len: %v off: %v", len(dest), off)

	return fuse.ReadResultData(f.Data[off:end]), fs.OK
}

func (f *MemRegularFile) Create(ctx context.Context, name string, flags uint32, mode uint32, out *fuse.EntryOut) (node *fs.Inode,
	fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {

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
