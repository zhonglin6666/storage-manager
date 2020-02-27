package file

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/sirupsen/logrus"
)

const (
	BLKSIZE = 4096

	DISKSIZE = 2
)

type FileSystem struct {
	Debug     bool
	WorkSpace string
	root      *Root
}

func NewFileSystem(workSpace string, debug bool) *FileSystem {
	return &FileSystem{
		Debug:     debug,
		WorkSpace: workSpace,
		root:      &Root{Ino: 2, WorkSpace: workSpace},
	}
}

func (m *FileSystem) Create() {
	logrus.Infof("FileSystem create workspace: %s begin", m.WorkSpace)

	opts := &fs.Options{}
	opts.Debug = m.Debug

	// TODO remove to create volume function
	if _, err := os.Stat("/tmp/file.img"); err == nil {
		f, _ := m.root.OpenDisk("/tmp/file.img", DISKSIZE)
		defer f.Close()

		metadataBytes := make([]byte, BLKSIZE)
		ReadBlock(f, 0, &metadataBytes)

		metadataMap := make(map[string]interface{})
		json.Unmarshal(metadataBytes, &metadataMap)

		fmt.Printf(" block- metadataMap: %#v", metadataMap)

		logrus.Infof(" Block-0 meta")

		metablock := make([]byte, BLKSIZE)
		ReadBlock(f, 1, &metablock)
		m.root.mb.DiskToMeta(metablock)

		logrus.Infof("zzlin file system create metadatamap: %#v", metadataMap)
		// m.root := setupDir(metadataMap
	}

	server, err := fs.Mount(m.WorkSpace, m.root, opts)
	if err != nil {
		log.Fatalf("Mount fail: %v\n", err)
	}

	go server.Wait()

	logrus.Infof("FileSystem create workspace: %s end", m.WorkSpace)
}

//func setupRoot(m map[string]interface{}) Root {
//	var root Root
//	for key, value := range m {
//		if key == "Inode" {
//			inode, _ := value.(float64)
//			root.Inode = uint64(inode)
//			inodeCount++
//		} else if key == "Name" {
//			dir.Name, _ = value.(string)
//		} else if key == "Files" {
//			var files []*filesys.File
//			allFiles, ok := value.([]interface{})
//			if !ok {
//				dir.Files = nil
//				continue
//			}
//			for _, i := range allFiles {
//				val, _ := i.(map[string]interface{})
//				file := setupFile(val)
//				files = append(files, &file)
//			}
//			dir.Files = &files
//		} else if key == "Directories" {
//			var dirs []*filesys.Dir
//			allDirs, ok := value.([]interface{})
//			if !ok {
//				dir.Directories = nil
//				continue
//			}
//			for _, i := range allDirs {
//				val, _ := i.(map[string]interface{})
//				dirToAppend := setupDir(val)
//				dirs = append(dirs, &dirToAppend)
//			}
//			dir.Directories = &dirs
//		}
//	}
//	return dir
//}
//
//func setupFile(m map[string]interface{}) filesys.File {
//	var file filesys.File
//	for key, value := range m {
//		if key == "Inode" {
//			inode, _ := value.(float64)
//			file.Inode = uint64(inode)
//			inodeCount++
//		} else if key == "Name" {
//			file.Name, _ = value.(string)
//			/*} else if key == "Data" {
//			data, _ := value.(string)
//			file.Data, _ = base64.StdEncoding.DecodeString(data)*/
//		} else if key == "Blocks" {
//			var blocks []int
//			allBlocks, ok := value.([]interface{})
//			if !ok {
//				file.Blocks = nil
//				continue
//			}
//			for _, i := range allBlocks {
//				val, _ := i.(float64)
//				blocks = append(blocks, int(val))
//			}
//			file.Blocks = blocks
//		} else if key == "Size" {
//			size, _ := value.(float64)
//			file.Size = uint64(size)
//		}
//	}
//	return file
//}
