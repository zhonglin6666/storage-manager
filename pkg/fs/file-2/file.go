package file_2

import (
	"math"
	"sync"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"bazil.org/fuse/fuseutil"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

type File struct {
	Node
	sync.Mutex

	dir *Dir
}

func (file *File) Attr(ctx context.Context, attr *fuse.Attr) error {
	logrus.Infof("  file Attributes for file: %s", file.Name)
	attr.Inode = file.Inode
	attr.Mode = 0777
	attr.Size = file.Size
	attr.BlockSize = uint32(BLKSIZE)
	attr.Blocks = uint64(len(file.Blocks))
	return nil
}

/* Look at this function later because it's not supposed to return the whole data */
func (file *File) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	logrus.Infof("  file Requested Read on File: %s", file.Name)
	data := make([]byte, 0, 10)

	for i := 0; i < len(file.Blocks); i++ {
		blockData := make([]byte, BLKSIZE)
		ReadBlock(file.dir.fi, file.Blocks[i], &blockData)
		data = append(data, blockData...)
	}
	fuseutil.HandleRead(req, resp, data)
	return nil
}

func (file *File) ReadAll(ctx context.Context) ([]byte, error) {
	logrus.Infof("  file Reading all of file: %s", file.Name)
	data := make([]byte, 0, 10)

	for i := 0; i < len(file.Blocks); i++ {
		blockData := make([]byte, BLKSIZE)
		ReadBlock(file.dir.fi, file.Blocks[i], &blockData)
		data = append(data, blockData...)
	}
	return []byte(data), nil
}

func (file *File) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	logrus.Infof("   File Trying to write to %s, data: %s", file.Name, string(req.Data))
	file.Lock()
	defer file.Unlock()

	resp.Size = len(req.Data)
	file.Size = uint64(len(req.Data))
	numBlocks := int(math.Ceil(float64(file.Size) / float64(BLKSIZE)))
	blocks := make([]int, numBlocks)

	k, allocatedBlocks := 0, 0
	for i := 0; i < numBlocks; i++ {
		var blocknr int
		if allocatedBlocks < len(file.Blocks) {
			blocknr = file.Blocks[i]
			allocatedBlocks++
		} else {
			blocknr = GetLowestFreeBlock()
		}
		var data []byte
		if i == numBlocks-1 {
			data = req.Data[k:]
		} else {
			data = req.Data[k:BLKSIZE]
		}
		k += BLKSIZE
		nbytes, err := WriteBlock(file.dir.fi, blocknr, data)
		logrus.Infof("BYTES WRITTEN: %d, err: %v", nbytes, err)
		blocks[i] = blocknr
	}
	file.Blocks = blocks
	file.dir.needUpdate <- struct{}{}

	logrus.Infof(" file Wrote to file: %s", file.Name)
	return nil
}

func (file *File) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	logrus.Infof(" file Flushing file: %s", file.Name)
	return nil
}

func (file *File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	logrus.Infof("  file Open call on file: %s", file.Name)
	return file, nil
}

func (file *File) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	logrus.Infof("  file Release requested on file: %s", file.Name)
	return nil
}

func (file *File) Fsync(ctx context.Context, req *fuse.FsyncRequest) error {
	logrus.Infof("  file Fsync requested on file: %s", file.Name)
	return nil
}
