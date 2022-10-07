//go:build windows

package main

import (
	"errors"
	"fmt"
	"golang.org/x/sys/windows"

	//	"fmt"
	//	"io"
	"os"
	"syscall"
	"time"
	//	"time"
	//	"github.com/djherbis/times"
	winio "github.com/Microsoft/go-winio"
)

type FileInfo struct {
	IsDir bool
	Atime time.Time `json:"atime"` // access time
	Ctime time.Time `json:"ctime"` // create time
	Mtime time.Time `json:"mtime"` // modified time
	Attrs uintptr   `json:"attrs"`
	Size  int64     `json:"size"`
}

func toTime(t windows.Filetime) time.Time {
	return time.Unix(0, t.Nanoseconds())
}

func openFile(path string) (*os.File, error) {
	pathp, e := syscall.UTF16PtrFromString(path)

	if e != nil {
		return nil, e
	}
	h, e := syscall.CreateFile(pathp,
		syscall.FILE_WRITE_ATTRIBUTES, syscall.FILE_SHARE_WRITE, nil,
		syscall.OPEN_EXISTING, syscall.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if e != nil {
		return nil, e
	}
	//	defer syscall.Close(h)
	f := os.NewFile(uintptr(h), path)
	return f, nil
}

func GetFileInfo(path string) (FileInfo, error) {
	var ret FileInfo
	osInfo, _ := os.Stat(path)
	f, err := openFile(path)
	defer f.Close()

	if err == nil {
		info, err := winio.GetFileBasicInfo(f)
		if err != nil {
			fmt.Println(err)
		}
		ret.IsDir = osInfo.IsDir()
		ret.Atime = toTime(info.LastAccessTime)
		ret.Ctime = toTime(info.CreationTime)
		ret.Mtime = osInfo.ModTime()
		ret.Attrs = uintptr(info.FileAttributes)
		//		fmt.Println(ret.Ctime)
		//		a, _ := times.StatFile(f)
		//		ret.Ctime = a.BirthTime()
		//		ret.Atime = a.AccessTime()
		//		fmt.Println(a.BirthTime())
		//		fmt.Println(info.FileAttributes)
		ret.Size = osInfo.Size()
	}
	return ret, nil
}

func SyncFileTime(src, dest string) {
	srcInfo, err := GetFileInfo(src)
	if err == nil {
		f, _ := openFile(dest)
		defer f.Close()
		info := winio.FileBasicInfo{}
		info.FileAttributes = uint32(srcInfo.Attrs)
		info.CreationTime = windows.NsecToFiletime(srcInfo.Ctime.UnixNano())
		info.ChangeTime = windows.NsecToFiletime(srcInfo.Mtime.UnixNano())
		winio.SetFileBasicInfo(f, &info)
		os.Chtimes(dest, srcInfo.Atime, srcInfo.Mtime)
	}
}

func CopyFileItem(src *FileItem, dpath string) (err error) {
	dInfo, err := GetFileInfo(dpath)

	if src.Info.IsDir && dInfo.IsDir {
		return nil
		//return errors.New("same file")
	}

	if src.Info.Size == dInfo.Size {
		dHash := GetFileHash(dpath)
		if src.GetHash() == dHash {
			fmt.Printf("%20x %20x    %-40s\n", src.GetHash(), dHash, dpath)
			return errors.New("same file")
		}
	}

	err = CopyFile(src.GetAbsPath(), dpath)
	if err == nil {
		SyncFileTime(src.GetAbsPath(), dpath)
	} else {
		fmt.Println("panic: ", src.GetAbsPath(), dpath)
		os.Exit(-1)
	}
	return err
}
