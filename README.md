

# diskText

- 直接跳过文件系统，对磁盘数据进行读写，判断磁盘测试数据读写校验数据是否丢失

##### 磁盘信息

```go
type DiskSizeInfo struct {
	DiskPath  string             // 磁盘名
	SrcMap    map[string]string  // 测试数据集(文件夹/校验数据)
	Size      int                // 磁盘分组大小
	SeekSize  int				 // 磁盘数据偏移
	BlockSize int                // 读写磁盘数据的大小
}
```

##### 使用两种方式跳过文件系统直接对磁盘进行读写

1. 直接读写磁盘

```go
// 磁盘的写入
func (d *DiskSizeInfo)DiskWriteByFile(src string) error
// 读取磁盘
func (d *DiskSizeInfo)DiskReadByFile() ([]byte, error)
```

​	2.通过dd命令行对磁盘进行读者

```go
// 通过dd命令写于磁盘
func (d *DiskSizeInfo)WriteDisk(staPos int, iFile string) error 
// 通过dd命令读取磁盘数据
func (d *DiskSizeInfo)ReadDisk(staPos int, oFile string) error
```

##### 对磁盘已有数据进行校验

```go
func (d *DiskSizeInfo)CheckReadDisk(src string, i int)
```

##### 磁盘校验数据的类型的返回

```go
func (d *DiskSizeInfo)DiskDatatype( i int) 
```

##### 循环通过数据集进项校验

```go
func (d *DiskSizeInfo)Run() error 
```

##### 初始化校验集数据文件

```go
func (d *DiskSizeInfo)initFile()error
```

##### 磁盘的挂起或者卸载（针对与带文件系统）和磁盘清除

```go
func (d *DiskSizeInfo)MountDisk(src string) error
func (d *DiskSizeInfo)ClearDisk(src string) error 
```

##### 初始化校验集数据文件

```go
func (d *DiskSizeInfo)initFile()
```

##### 其他功能

```go
文件通过md5校验比较
func FileCompare(src, dst string) (bool, error)
生成文件校验和函数
func fileCheckSum(fileName string) (string, error)
//PathExists 判断一个文件或文件夹是否存在
//输入文件路径，根据返回的bool值来判断文件或文件夹是否存在
func PathExists(path string) (bool,error)
校验文件中是否存在校验值
func CheckFileData(data []byte, check string) bool
```
