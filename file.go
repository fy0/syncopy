package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spaolacci/murmur3"
)

type FileItem struct {
	Path   string       `json:"path"`
	Info   FileInfo     `json:"info"`
	Hash   uint64       `json:"hash"`
	parent *FileSummary `json:"parent"`
}

type FileItemSlice []FileItem

func (items FileItemSlice) Len() int {
	return len(items)
}

func (items FileItemSlice) Swap(i, j int) {
	items[i], items[j] = items[j], items[i]
}

func (items FileItemSlice) Less(i, j int) bool {
	return items[i].Path < items[j].Path
}

func (item *FileItem) GetAbsPath() string {
	r, _ := filepath.Abs(filepath.Join(item.parent.Path, item.Path))
	return r
}

func (item *FileItem) GetHash() uint64 {
	if item.Hash == 0 {
		item.Hash = GetFileHash(item.GetAbsPath())
	}
	return item.Hash
}

type FileSummary struct {
	Path  string        `json:"path"`
	Time  time.Time     `json:"time"`
	Items FileItemSlice `json:"items"`
}

func GetFileHash(fn string) uint64 {
	fp, err := os.Open(fn)
	if err == nil {
		buf := make([]byte, 1048576)
		fp.Read(buf)
		h := murmur3.New64()
		h.Write(buf)
		return h.Sum64()
	}
	return 0
}

func CopyFile(src, dst string) (err error) {
	sfi, err := os.Stat(src)
	if err != nil {
		return
	}
	if sfi.IsDir() {
		os.MkdirAll(dst, sfi.Mode())
		return
	}
	if !sfi.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
		if os.SameFile(sfi, dfi) {
			return
		}
	}
	if err = os.Link(src, dst); err == nil {
		return
	}
	err = copyFileContents(src, dst)
	return
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}
