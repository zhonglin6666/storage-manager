package file_2

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"

	fsinterface "storage-manager/pkg/fs"
)

var inodeCount uint64 = 10

type FS struct {
	RootDir *Dir

	Debug     bool
	WorkSpace string

	needUpdate chan struct{}
}

func (f *FS) Root() (fs.Node, error) {
	return f.RootDir, nil
}

func (f *FS) Destroy() {}

func NewFileSystem(workSpace string, debug bool) fsinterface.FileSystem {
	f := &FS{
		Debug:      debug,
		WorkSpace:  workSpace,
		RootDir:    &Dir{},
		needUpdate: make(chan struct{}),
	}

	if _, err := os.Stat("sda"); err == nil {
		fi, _ := OpenDisk("sda", DISKSIZE)
		defer fi.Close()

		metadataBytes := make([]byte, BLKSIZE)
		ReadBlock(fi, 0, &metadataBytes)

		metadataMap := make(map[string]interface{})
		json.Unmarshal(metadataBytes, &metadataMap)

		fmt.Printf(" block- metadataMap: %#v", metadataMap)

		metablock := make([]byte, BLKSIZE)
		ReadBlock(fi, 1, &metablock)
		DiskToMeta(metablock)

		logrus.Infof("zzlin file system create metadatamap: %#v", metadataMap)
		*(f.RootDir) = setupDir(metadataMap)
	}

	f.RootDir.needUpdate = make(chan struct{})

	return f
}

func (f *FS) Create() {
	logrus.Infof("FileSystem create workspace: %s begin", f.WorkSpace)

	conn, err := fuse.Mount(f.WorkSpace)
	if err != nil {
		logrus.Errorf("fuse mount err: %v", err)
		return
	}
	defer conn.Close()

	server := fs.New(conn, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go waitForUpdateMetada(f.RootDir, ctx)

	if err := server.Serve(f); err != nil {
		logrus.Errorf("server serve err: %v", err)
		return
	}

	<-conn.Ready
	if err := conn.MountError; err != nil {
		logrus.Errorf("conn mount error: %v", err)
	}

	logrus.Infof("FileSystem create workspace: %s end", f.WorkSpace)
}

func waitForUpdateMetada(d *Dir, ctx context.Context) {
	fi, _ := OpenDisk("sda", DISKSIZE)
	defer fi.Close()
	d.fi = fi

	for {
		select {
		case <-ctx.Done():
			logrus.Infof("waitForUpdateMetada quit")
			goto over
		case <-d.needUpdate:
			MetaToDisk(fi)
			WriteFSToDisk(fi, d)
		}
	}
over:
}

func setupDir(m map[string]interface{}) Dir {
	var dir Dir
	for key, value := range m {
		if key == "Inode" {
			inode, _ := value.(float64)
			dir.Inode = uint64(inode)
			inodeCount++
		} else if key == "Name" {
			dir.Name, _ = value.(string)
		} else if key == "Files" {
			var files []*File
			allFiles, ok := value.([]interface{})
			if !ok {
				dir.Files = nil
				continue
			}
			for _, i := range allFiles {
				val, _ := i.(map[string]interface{})
				file := setupFile(val)
				files = append(files, &file)
			}
			dir.Files = &files
		} else if key == "Directories" {
			var dirs []*Dir
			allDirs, ok := value.([]interface{})
			if !ok {
				dir.Directories = nil
				continue
			}
			for _, i := range allDirs {
				val, _ := i.(map[string]interface{})
				dirToAppend := setupDir(val)
				dirs = append(dirs, &dirToAppend)
			}
			dir.Directories = &dirs
		}
	}
	return dir
}

func setupFile(m map[string]interface{}) File {
	var file File
	for key, value := range m {
		if key == "Inode" {
			inode, _ := value.(float64)
			file.Inode = uint64(inode)
			inodeCount++
		} else if key == "Name" {
			file.Name, _ = value.(string)
		} else if key == "Blocks" {
			var blocks []int
			allBlocks, ok := value.([]interface{})
			if !ok {
				file.Blocks = nil
				continue
			}
			for _, i := range allBlocks {
				val, _ := i.(float64)
				blocks = append(blocks, int(val))
			}
			file.Blocks = blocks
		} else if key == "Size" {
			size, _ := value.(float64)
			file.Size = uint64(size)
		}
	}
	return file
}
