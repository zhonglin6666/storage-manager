package file

import (
	"bytes"
	"encoding/json"
	"os"

	"github.com/Workiva/go-datastructures/bitarray"
	"github.com/sirupsen/logrus"
)

type MetaBlock struct {
	Bitmap     []byte
	LowestFree int
}

func NewBlocks(fd *os.File, size, blksize uint64) *MetaBlock {
	ba := bitarray.NewBitArray(size / blksize)
	ba.SetBit(0)
	ba.SetBit(1)
	data, _ := bitarray.Marshal(ba)

	return &MetaBlock{
		Bitmap:     data,
		LowestFree: 2,
	}
}

func ReadBlock(disk *os.File, blocknr int, data *[]byte) (int, error) {
	if _, err := disk.Seek(int64(blocknr*BLKSIZE), 0); err != nil {
		return 0, err
	}
	nbytes, err := disk.Read(*data)
	if err != nil {
		return 0, err
	}
	*data = bytes.Trim(*data, string(byte(0)))
	return nbytes, nil
}

func (mb *MetaBlock) WriteRootToDisk(f *os.File, r *Root) {
	logrus.Infof("     WriteRootToDisk root: %#v", r)
	rootBytes, _ := json.Marshal(r)
	mb.WriteBlock(f, 0, rootBytes)
	logrus.Infof(" write root info to disk")
}

func (mb *MetaBlock) MetaToDisk(f *os.File) {
	metablock, _ := json.Marshal(*mb)
	mb.WriteBlock(f, 1, metablock)
	logrus.Infof("  write meta to disk")
}

func (mb *MetaBlock) DiskToMeta(data []byte) {
	json.Unmarshal(data, &mb)
}

func (mb *MetaBlock) WriteBlock(disk *os.File, blocknr int, data []byte) (int, error) {
	zeros := make([]byte, BLKSIZE)
	if _, err := disk.Seek(int64(blocknr*BLKSIZE), 0); err != nil {
		return 0, err
	}
	disk.Write(zeros)

	if len(data) == 0 {
		updateBlocks(mb, "del", blocknr)
		return 0, nil
	}
	if _, err := disk.Seek(-int64(BLKSIZE), 1); err != nil {
		return 0, err
	}

	nbytes, err := disk.Write(data)
	if err != nil {
		return 0, err
	}

	updateBlocks(mb, "set", blocknr)
	return nbytes, nil
}

func updateBlocks(mb *MetaBlock, operation string, blocknr int) {
	if operation == "del" {
		ba, _ := bitarray.Unmarshal(mb.Bitmap)
		ba.ClearBit(uint64(blocknr))
		data, _ := bitarray.Marshal(ba)
		mb.Bitmap = data
		if blocknr < mb.LowestFree {
			mb.LowestFree = blocknr
		}
	} else if operation == "set" {
		ba, _ := bitarray.Unmarshal(mb.Bitmap)

		if ok, _ := ba.GetBit(uint64(blocknr)); !ok {
			ba.SetBit(uint64(blocknr))
			data, _ := bitarray.Marshal(ba)
			mb.Bitmap = data
			if blocknr == mb.LowestFree {
				i := blocknr + 1
				for i < int(ba.Capacity()) {
					if ok, _ := ba.GetBit(uint64(i)); !ok {
						mb.LowestFree = i
						break
					}
					i++
				}
			}
		}
	}

	logrus.Infof("New lowest free block:", mb.LowestFree)
}
