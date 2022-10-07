// syncode project main.go
package main

import (
	"fmt"
	"io/fs"
	"os"
	"sort"
	"time"

	"encoding/json"
	"path/filepath"

	"gopkg.in/urfave/cli.v1"
	//	"github.com/djherbis/times"
	//	"github.com/fsnotify/fsnotify"
)

func LoadIgnore(dirPath string) *SyncIgnore {
	ret := new(SyncIgnore)
	fn := filepath.Join(dirPath, ".scignore")
	_, err := os.Stat(fn)

	if !os.IsNotExist(err) {
		if ret.ReadFile(fn) != nil {
			return nil
		}
		return ret
	}

	return nil
}

func Scan(dirPath string) FileSummary {
	var lst []FileItem
	var ret FileSummary
	ignore := LoadIgnore(dirPath)
	if ignore != nil {
		fmt.Printf(" -> scignore loaded")
	}

	_ = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		p, _ := filepath.Rel(dirPath, path)
		if p != "." {
			i, _ := GetFileInfo(path)
			if ignore != nil {
				m := ignore.Match(p)
				//m := false
				if !m {
					lst = append(lst, FileItem{p, i, 0, &ret})
				} else {
					//fmt.Printf("ignore: %s\n", p)
					if i.IsDir {
						return fs.SkipDir
					}
				}
			} else {
				lst = append(lst, FileItem{p, i, 0, &ret})
			}
		}
		return nil
	})

	ret.Time = time.Now()
	ret.Path = dirPath
	ret.Items = lst
	return ret
}

func main() {
	app := cli.NewApp()
	app.Name = "syncode"
	app.Usage = "file sync"

	srcPathLst := cli.StringSlice{}
	destPathLst := cli.StringSlice{}

	app.Flags = []cli.Flag{
		//cli.StringFlag{
		//	Name:  "config, c",
		//	Usage: "Load configuration from `FILE`",
		//},
		cli.StringSliceFlag{
			Name:  "src, s",
			Usage: "Path for src directories",
			Value: &srcPathLst,
		},
		cli.StringSliceFlag{
			Name:  "dest, d",
			Usage: "Path for dest directories",
			Value: &destPathLst,
		},
	}

	app.Action = func(c *cli.Context) error {
		if srcPathLst.String() == "[]" || destPathLst.String() == "[]" {
			cli.ShowAppHelp(c)
			return nil
		}

		fmt.Println("原始路径: ", srcPathLst)
		fmt.Println("目标路径: ", destPathLst)

		fmt.Println("")
		fmt.Println("摘要生成")

		var scanResults []FileSummary

		for i, dirPath := range srcPathLst {
			dp, _ := filepath.Abs(dirPath)
			fmt.Printf("%2d. %s", i+1, dp)
			s := Scan(dirPath)
			fmt.Printf(" -> 文件 %d 个\n", len(s.Items))
			scanResults = append(scanResults, s)
		}

		getNewest := func(items []FileItem) (*FileItem, []string) {
			var recent *FileItem
			var recentTime int64

			if len(items) == 1 {
				return &items[0], []string{}
			}

			conflict := []string{}
			uniqueItems := []FileItem{}

			for _, i := range items {
				for _, j := range uniqueItems {
					if i.Info.IsDir && j.Info.IsDir {
						//						goto loop_end
					}
					if i.Info.Size == j.Info.Size {
						if i.GetHash() == j.GetHash() {
							goto loop_end
						}
					}
				}
				uniqueItems = append(uniqueItems, i)
			loop_end:
			}

			if len(uniqueItems) > 1 {
				for _, i := range uniqueItems {
					conflict = append(conflict, i.Path)
				}
			}

			for _, i := range uniqueItems {
				t := i.Info.Mtime.Unix()
				if t > recentTime {
					recent = &i
					recentTime = t
				}
			}

			return recent, conflict
		}

		fmt.Println("\n摘要合并")

		makeTree := func(scanResults []FileSummary) FileSummary {
			var s FileSummary
			fnLst := make(map[string]([]FileItem))

			for _, i := range scanResults {
				for _, j := range i.Items {
					el, ok := fnLst[j.Path]
					if ok {
						fnLst[j.Path] = append(el, j)
					} else {
						fnLst[j.Path] = []FileItem{j}
					}
				}
			}

			for k, v := range fnLst {
				item, conflict := getNewest(v)
				if len(conflict) > 0 {
					p, _ := filepath.Abs(item.parent.Path)
					fmt.Printf(" 冲突：%s，取目录 %s 版本\n", k, p)
				}
				s.Items = append(s.Items, *item)
			}

			sort.Sort(&s.Items)
			//			for _, i := range s.Items {
			//				fmt.Println(i.Path)
			//			}
			s.Time = time.Now()
			return s

		}
		summary := makeTree(scanResults)

		jsonObj, err := json.Marshal(summary)
		if err != nil {
			fmt.Println("err: ", err)
			os.Exit(-2)
		} else {
			f, _ := os.Create("_files.json")
			f.Write(jsonObj)
			f.Close()
		}

		fmt.Println("\n数据同步")

		// 先创建目录
		dirMap := map[string]string{}
		for _, srcRoot := range srcPathLst {
			for _, destRoot := range destPathLst {
				dirMap[srcRoot] = destRoot
			}
		}

		for _, si := range summary.Items {
			for _, destRoot := range destPathLst {
				if si.Info.IsDir {
					dest := filepath.Join(destRoot, si.Path)
					CopyFileItem(&si, dest)
					dirMap[si.GetAbsPath()] = dest
				}
			}
		}

		// 复制文件
		synched := []string{}
		for _, si := range summary.Items {
			//fmt.Println(si.GetAbsPath(), filepath.Join(i.Path, si.Path))
			for _, destRoot := range destPathLst {
				if !si.Info.IsDir {
					err := CopyFileItem(&si, filepath.Join(destRoot, si.Path))
					if err == nil {
						synched = append(synched, filepath.Join(destRoot, si.Path))
					} else if err.Error() == "same file" {
						// fmt.Println(err.Error())
					}
				}
			}
		}
		fmt.Printf("同步文件 %d 个\n", len(synched))
		for _, i := range synched {
			fmt.Printf("  %s\n", i)
		}

		// 刷新目录修改时间
		for k, v := range dirMap {
			SyncFileTime(k, v)
		}

		return nil
	}
	app.Run(os.Args)

	/* fmt.Println("Hello World!")
	fmt.Println("当前时间", time.Now()).\main.go:38:20: cannot use t (type time.Time) as type syscall.Filetime in assignment
	fmt.Println(add(1, 3))
	fmt.Println(fib(30))
	fmt.Println(fib(10))
	fmt.Println(Sqrt(10))*/
}
