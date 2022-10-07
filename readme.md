sync copy
----------------

将目录src的文件同步到目录dest，跳过相同文件且保持时间戳不变。

用于备份或迁移数据，存旧如旧。

仅限于Windows。


Sync files from directory src to directory dest, skip same files and preserve all timestamps.

Used for backing up or migrating files. Keep everything in the past.

Only for Windows.


```cmd
> scopy.exe -s src -d dest

原始路径:  [src]
目标路径:  [dest]

摘要生成
 1. C:\work\syncopy\1 -> scignore loaded -> 文件 15 个

摘要合并

数据同步
    c6c45e8226f29942     c6c45e8226f29942    dest\.scignore
    d1f583d2fb3eb7a9     d1f583d2fb3eb7a9    dest\11.txt
    d1f583d2fb3eb7a9     d1f583d2fb3eb7a9    dest\123\新建文本文档.txt
    d1f583d2fb3eb7a9     d1f583d2fb3eb7a9    dest\4\3.txt
    c763d318cbe77007     c763d318cbe77007    dest\ddd.go                             
    abdc54c58c081d10     abdc54c58c081d10    dest\main.go
    d1f583d2fb3eb7a9     d1f583d2fb3eb7a9    dest\新建文本文档 (2).txt
同步文件 4 个
  dest\22.txt
  dest\doc.go
  dest\新建 Microsoft Word 文档.docx
  dest\新建文本文档.txt

```

另外，可以使用`.scignore`屏蔽你不想要的文件，规则使用`github.com/gobwas/glob`
