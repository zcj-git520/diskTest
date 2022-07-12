/*
	author：zhaochengji mail：909536346@qq.com
	// 直接跳过文件系统，对磁盘数据进行读写
	磁盘测试数据读写校验数据是否丢失
*/
package diskTest

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"github.com/ncw/directio"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

const (
	BLOCKSIZE = 4096 // 块设备读写的最少单位
	/*内存单位*/
	B = 1
	KB = 1024 * B
	MB = 1024 * KB
	G = 1000 * MB
)

// 校验磁盘的信息
type DiskSizeInfo struct {
	DiskPath  string             // 磁盘名
	SrcMap    map[string]string  // 测试数据集(文件夹/校验数据)
	Size      int                // 磁盘分组大小
	SeekSize  int				 // 磁盘数据偏移
	BlockSize int                // 读写磁盘数据的大小
}

// 磁盘的写入
func (d *DiskSizeInfo)DiskWriteByFile(src string) error {
	_, err := os.Stat(src)
	if err != nil {
		return  err
	}
	// 跳过文件系统缓存，直接读写磁盘
	// 打开输入文件数据
	source, err := directio.OpenFile(src, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	defer source.Close()
	// 获取磁盘的数据的块的大小的偏移
	buf := directio.AlignedBlock(d.BlockSize)
	// 打开磁盘
	destination, err :=  directio.OpenFile(d.DiskPath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer destination.Close()
	nBytes, err := source.Read(buf)
	if err != nil {
		return  err
	}
	// 对磁盘进行偏移，并写于数据到磁盘
	_, _ = destination.Seek(int64(d.SeekSize), io.SeekStart)
	nBytes, err = destination.Write(buf)
	if err != nil {
		log.Println(nBytes)
		return err
	}
	return  err
}

// 读取磁盘
func (d *DiskSizeInfo)DiskReadByFile() ([]byte, error) {
	_, err := os.Stat(d.DiskPath)
	if err != nil {
		return  nil, err
	}
	source, err := directio.OpenFile(d.DiskPath, os.O_RDONLY, 0666)
	if err != nil {
		return  nil, err
	}
	defer source.Close()
	buf := directio.AlignedBlock(d.BlockSize)
	_, err = source.Seek(int64(d.SeekSize), io.SeekStart)
	if err != nil {
		log.Println("seek", d.SeekSize)
		return  nil, err
	}
	_, err = source.Read(buf)
	if err != nil {
		return  nil, err
	}
	return  buf, err
}

// 磁盘的挂起或者卸载（针对与带文件系统）
func (d *DiskSizeInfo)MountDisk(src string) error {
	// 对磁盘的挂起或者卸载
	// 卸载磁盘
	Str := fmt.Sprintf("umount %s", d.DiskPath)
	cmd := exec.Command("/bin/sh", "-c", Str)
	err := cmd.Run()
	if err != nil {
		log.Println("umount:",err.Error())
		return err
	}
	// 格式化磁盘
	Str = fmt.Sprintf("mkfs.ext4 %s", d.DiskPath)
	cmd = exec.Command("/bin/sh", "-c", Str)
	err = cmd.Run()
	if err != nil {
		log.Println("mkfs:", err.Error())
		return err
	}
	mountDir := "./mountData"
	ok, _ := PathExists(mountDir)
	if !ok{
		 _ = os.Mkdir(mountDir, os.ModePerm)
	}
	// 挂起磁盘
	Str = fmt.Sprintf("mount %s %s", d.DiskPath, mountDir)
	cmd = exec.Command("/bin/sh", "-c", Str)
	err = cmd.Run()
	if err != nil {
		log.Println("mount:",err.Error())
		return err
	}
	return nil
}

// clear disk
func (d *DiskSizeInfo)ClearDisk(src string) error {
	AllSize := d.DiskSize()
	bk := AllSize / d.Size
	for i:=0; i < bk; i++{
		err := d.DiskWriteByFile(src)
		if err != nil {
			fmt.Println(err.Error(), i, bk)
			return err
		}
	}
	return nil
}

// 获取磁盘的大小
func (d *DiskSizeInfo)DiskSize()  int {
	Str := fmt.Sprintf("df -h %s", d.DiskPath)
	cmd := exec.Command("/bin/sh", "-c", Str)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Println(err.Error())
		return 0
	}
	str := out.String()
	log.Println(str)
	// 获取磁盘有用的空间大小
	re := regexp.MustCompile("[0-9]+")
	data := re.FindAllString(str, -1)
	size, _ := strconv.Atoi(data[2])
	sizeType := str[len(str)-1:]
	if sizeType == "G" {
		return size * G
	}else if sizeType == "M"{
		return size * MB
	}
	return size * KB
}

// 通过dd命令写于磁盘
func (d *DiskSizeInfo)WriteDisk(staPos int, iFile string) error {
	ok, err := PathExists(iFile)
	if !ok{
		return err
	}
	Str := fmt.Sprintf("dd if=%s of=%s bs=%dM count=%d seek=%d iflag=direct", iFile, d.DiskPath, d.Size, 1, staPos)
	cmd := exec.Command( "/bin/sh", "-c", Str)
	err = cmd.Run()
	if err != nil {
		log.Println("write fail:", err.Error(), staPos)
		return err
	}
	return nil
}

// 通过dd命令读取磁盘数据
func (d *DiskSizeInfo)ReadDisk(staPos int, oFile string) error {
	ok, err := PathExists(d.DiskPath)
	if !ok{
		return err
	}
	Str := fmt.Sprintf("dd if=%s of=%s bs=%dM count=%d skip=%d iflag=direct", d.DiskPath, oFile, d.Size,1, staPos)
	cmd := exec.Command( "/bin/sh", "-c", Str)
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// 对磁盘已有数据进行校验
func (d *DiskSizeInfo)CheckReadDisk(src string, i int) error {
	log.Printf("******************************************Check %s\n", src)
	tmpFile := "./tem.txt"
	ok, _:= PathExists(tmpFile)
	if ok {
		_ = os.RemoveAll(tmpFile)
	}
	data, err := d.DiskReadByFile()
	if err != nil {
		return err
	}
	err = saveFile(tmpFile, data)
	if err != nil {
		return err
	}
	ok, err = FileCompare(src, tmpFile)
	if err != nil {
		return err
	}
	if !ok{
		_ = os.RemoveAll(tmpFile)
		return fmt.Errorf("different file")

	}
	_ = os.RemoveAll(tmpFile)
	return nil
}

// 磁盘校验数据的类型的返回
func (d *DiskSizeInfo)DiskDatatype( i int) string {
	for src, check := range d.SrcMap {
		data, err := d.DiskReadByFile()
		if err != nil {
			fmt.Println(err.Error())
			return ""
		}
		if CheckFileData(data, check) {
			return src
		}
	}
	return ""
}

// 循环通过数据集进项校验
func (d *DiskSizeInfo)Run() error {
	tmpFile := "./tmp1.txt"
	AllSize := d.DiskSize()
	bk := AllSize / d.Size  // 对磁盘进项分组
	if bk < 1{
		return fmt.Errorf("small disk data, size is:%d", AllSize)
	}
	fmt.Println("disk size: ", AllSize, bk)
	num := 0
	for {
		for src, check := range d.SrcMap {
			for i := 0; i < bk; i++ {
				// 对已有磁盘数据进行校验
				secType := d.DiskDatatype(i)
				if secType != "" {
					err := d.CheckReadDisk(secType, i)
					if err != nil {
						return fmt.Errorf("different by check")
					}
				}
				log.Println("################################################################")
				ok, _ := PathExists(tmpFile)
				if ok {
					_ = os.RemoveAll(tmpFile)
				}
				// 磁盘数据
				log.Printf("read block %v \n", i)
				data, err := d.DiskReadByFile()
				if err != nil {
					log.Println(err.Error())
					return err
				}
				// 校验是为测试集中数据，不是则为空
				if !CheckFileData(data, check) {
					log.Printf("block  %v is empty \n", i)
					log.Printf("write block %v \n", i)
					// 写入数据
					err := d.DiskWriteByFile(src)
					if err != nil {
						log.Println(err.Error())
						return err
					}
					// 写入数据后在之前磁盘进行校验
					i = -1
					continue
				}
				err = saveFile(tmpFile, data)
				if err != nil {
					return err
				}
				// 与测试集数据进行验证
				ok, err = FileCompare(src, tmpFile)
				if err != nil {
					return err
				}
				if !ok {
					log.Printf("block %v is different for src\n", i)
					_ = os.RemoveAll(tmpFile)
					return err

				}
				log.Printf("block %v is same for src\n", i)
				_ = os.RemoveAll(tmpFile)
			}
			num++
			log.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!ok!!!!!!!!!!!!!!!!!!!!!!!,num:= ", num)
		}
	}
}

// 初始化校验集数据文件
func (d *DiskSizeInfo)initFile()error{
	if len(d.SrcMap) <= 0{
		return fmt.Errorf("no file on SrcMap")
	}
	for src, _ := range d.SrcMap {
		_, err := os.Stat(src)
		if err != nil {
			// 测试集文件不存在，则在map中移除
			delete(d.SrcMap, src)
			continue
		}
		source, err := directio.OpenFile(src, os.O_RDONLY, 0666)
		if err != nil {
			// 测试集文件打开失败，则在map中将其移除
			delete(d.SrcMap, src)
			continue
		}
		defer source.Close()
		// 通过块设备的大小在重新生成测试集文件
		buf := directio.AlignedBlock(d.BlockSize)
		_, err = source.Read(buf)
		if err != nil {
			// 测试集文件数据读取失败，则在map中将其移除
			delete(d.SrcMap, src)
			continue
		}
		err = saveFile(src, buf)
		if err != nil {
			// 测试集文件保存失败，则在map中将其移除
			delete(d.SrcMap, src)
			continue
		}
	}
	// 判断是否还存在测试集文件
	if len(d.SrcMap) <= 0{
		return fmt.Errorf("error file on SrcMap")
	}
	return nil
}

// 初始化磁盘校验
func InitDiskTest(diskPath string, srcMap map[string]string, size int, seekSize  int, blockSize int)(*DiskSizeInfo, error) {
	disk := &DiskSizeInfo{
		DiskPath:  diskPath,
		SrcMap:    srcMap,
		Size:      size,
		SeekSize:  seekSize,
		BlockSize: blockSize,
	}
	err := disk.initFile()
	if err != nil {
		return nil, err
	}
	return disk, nil
}

// 文件通过md5校验比较
func FileCompare(src, dst string) (bool, error){
	srcCheckSum, err := fileCheckSum(src)
	if err != nil {
		return false, err
	}
	dstCheckSum , err := fileCheckSum(dst)
	if err != nil {
		return false, err
	}
	//fmt.Println("srcCheckSum:", srcCheckSum)
	//fmt.Println("dstCheckSum:", dstCheckSum)
	if strings.Contains(srcCheckSum, dstCheckSum){
		return true, nil
	}
	return false, err
}

func saveFile(des string, data []byte) error {
	if len(data) <= 0 {
		return fmt.Errorf("no data")
	}
	f, err := os.Create(des)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err != nil {
		return err
	}
	return nil
}

// 生成文件校验和函数
func fileCheckSum(fileName string) (string, error)  {
	f, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := md5.New()
	//h := sha256.New()
	//h := sha1.New()
	//h := sha512.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	// 格式化为16进制字符串
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

//PathExists 判断一个文件或文件夹是否存在
//输入文件路径，根据返回的bool值来判断文件或文件夹是否存在
func PathExists(path string) (bool,error) {
	_,err := os.Stat(path)
	if err == nil {
		return true,nil
	}
	if os.IsNotExist(err) {
		return false,nil
	}
	return false,err
}

// 校验文件中是否存在校验值
func CheckFileData(data []byte, check string) bool{
	if len(data) <= 0{
		return false
	}
	if strings.Contains(string(data), check){
		return true
	}
	return false
}