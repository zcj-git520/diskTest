/*
	author：zhaochengji mail：909536346@qq.com
	// 直接跳过文件系统，对磁盘数据进行读写
	磁盘测试数据读写校验数据是否丢失
*/
package diskTest

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"github.com/ncw/directio"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"sysd_log"
	"time"
	"timerTask"
)

func init()  {
	rand.Seed(time.Now().UnixNano())
}

var sourceBufData []byte  // 检测源数据
var Data []byte           // 随机生成数
var  LogNew = sysd_log.SysLogInit("./", "testDisk.log")
var log = LogNew.Logger
const (
	TimerTime = 2  // 定时时间
	readBlockCount = 10000 // 读磁盘块数
	writeBlockCount = 10   // 写磁盘块数
	BLOCKSIZE = 4096 // 块设备读写的最少单位
	/*内存单位*/
	B  = 1
	KB = 1024 * B
	MB = 1000 * KB
	G  = 1000 * MB
)

// 校验磁盘的信息
type DiskSizeInfo struct {
	DiskPath  		string             // 磁盘名
	Size      		float64                // 磁盘大小
	SeekSize  		int				 	// 磁盘数据偏移
	BlockSize 		int                // 读写磁盘数据的大小
	IsPower      	bool
	OutPowerNum 	int
	diskContext     context.Context
	diskMutex       sync.Mutex
	powerInNum          int
	powerOutNum         int

}
var blockWriteSize int

// 随机生成数
func RandomString(n int) []byte {
	var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return b
}

//生成源数据
func sourceData()  {
	Data = RandomString(8)
	log.Infof("******************check data is: %v **************************", string(Data))
	if len(sourceBufData) <= 0{
		sourceBufData = make([]byte, writeBlockCount*BLOCKSIZE)
	}
	for i:= 0; i< writeBlockCount*BLOCKSIZE; i++{
		j := 0
		if i != 0{
			j = i % len(Data)
		}
		sourceBufData[i] = Data[j]
	}
}

// 磁盘的写入
func (d *DiskSizeInfo)DiskWriteByFile() error {
	if !d.DiskStatus(){
		if !d.IsPower{
			d.PowerIn()
			return nil
		}
		return nil
	}
	destination, err :=  directio.OpenFile(d.DiskPath, os.O_WRONLY, 0666)
	if err != nil {
		log.Errorf("open %v disk error:%v ",d.DiskPath, err)
		return err
	}
	defer destination.Close()
	// 获取磁盘的数据的块的大小的偏移
	//buf := directio.AlignedBlock(d.BlockSize)
	// 打开磁盘
	if float64(d.BlockSize * BLOCKSIZE + d.SeekSize) >= (d.Size - 10*MB){
		//d.BlockSize = 0
		//d.SeekSize = 0
		return fmt.Errorf("disk write is over!!!" )
	}
	// 对磁盘进行偏移，并写于数据到磁盘
	_, err = destination.Seek(int64(d.BlockSize * BLOCKSIZE + d.SeekSize), io.SeekStart)
	if err != nil {
		log.Infof("seek error: %v \n", err)
		return  err
	}
	n := d.SeekSize % len(Data)
	nBytes, err := destination.Write(sourceBufData[n:])
	if err != nil {
		d.SeekSize = nBytes
		log.Warnf("%v :write disk is error: %v, write in to disk is %v byte\n", d.DiskPath, err, d.SeekSize)
		return err
	}
	d.BlockSize += writeBlockCount
	writeSize := (d.BlockSize * BLOCKSIZE + d.SeekSize)/ MB
	if blockWriteSize != writeSize{
		blockWriteSize = writeSize
		log.Infof("%v : write disk is ok, total write size is %v MB\n", d.DiskPath, writeSize)
	}
	return nil
}

// 读取磁盘
func (d *DiskSizeInfo)DiskReadByFile(index int, size int) ([]byte, error) {
	_, err := os.Stat(d.DiskPath)
	if err != nil {
		log.Errorf(err.Error())
		return nil, err
	}
	source, err := directio.OpenFile(d.DiskPath, os.O_RDONLY, 0666)
	if err != nil {
		log.Errorf("open %v disk error:%v ",d.DiskPath, err)
		return nil, err
	}
	defer source.Close()
	_, err = source.Seek(int64(index * BLOCKSIZE) , io.SeekStart)
	if err != nil {
		log.Println("seek", d.SeekSize)
		return  nil, err
	}
	//buf := make([]byte, size)
	buf := directio.AlignedBlock(size)
	_, err = source.Read(buf)
	if err != nil {
		log.Warnf("%v: read disk is error: %v\n",d.DiskPath, err)
		return  nil, err
	}
	//log.Infof("%v: read disk is ok",d.DiskPath)
	return  buf, nil
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

// 获取磁盘的大小
func DiskSize(DiskPath string)  float64 {
	Str := fmt.Sprintf("lsblk %s", DiskPath)
	cmd := exec.Command("/bin/sh", "-c", Str)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Error("cmd is error: ",err.Error())
		return 0
	}
	strs := out.String()
	//log.Println(strs)
	strPaths := strings.Split(DiskPath, "/")
	strPath := strPaths[len(strPaths) -1]
	strDatas := strings.Split(strs, "\n")
	for _, strData := range strDatas{
		if strings.Contains(strData, strPath){
			// 获取磁盘有用的空间大小
			data := strings.Split(strData, " ")
			str := ""
			for _,s := range data{
				if strings.Contains(s, "G") || strings.Contains(s, "K") ||
					strings.Contains(s, "M") {
					str = s
					break
				}
			}
			if str == ""{
				log.Errorf("no find: %v\n", DiskPath)
				return  0
			}
			size, _ := strconv.ParseFloat(str[:(len(str)-1)], 64)
			sizeType := str[(len(str)-1):]
			log.Info("size is:",size, "  size Type is :", sizeType)
			if sizeType == "G" {
				return size * G
			}else if sizeType == "M"{
				return size * MB
			}
			return size * KB
		}
	}
	return  0
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

func (d *DiskSizeInfo)check(index int, size int)error {
	data, err := d.DiskReadByFile(index, size)
	if err != nil {
		return err
	}
	// 与测试集数据进行验证
	for i, by := range data {
		j := 0
		if i != 0 {
			j = i % len(Data)
		}
		if Data[j] != by {
			log.Panicf("%v byte %v is different:disk data is:%v, check data is :%v, 不一致后续8个byte的数据为:%v",
				d.DiskPath, i, string(by), string(Data[j]), string(data[i:i+8]))
		}
	}
	return nil
}
// 对磁盘已有数据进行校验
func (d *DiskSizeInfo)CheckReadDisk(ctx context.Context) error {
	if !d.DiskStatus(){
		if !d.IsPower{
			d.PowerIn()
			return nil
		}
		return nil
	}
	if d.BlockSize <= 0 {
		return  fmt.Errorf("size is zero")
	}
	//t := 500
	//n :=  d.BlockSize / t
	//if n <= 0{
	//	err := d.check(0, d.BlockSize*BLOCKSIZE+d.SeekSize, source)
	//	if err != nil {
	//		return err
	//	}
	//}else{
	//	for i := 0; i < n ; i++{
	//		size := t * BLOCKSIZE
	//		//if i == n - 1{
	//		//	size += d.SeekSize
	//		//}
	//		go func(index, size int, ctx context.Context,){
	//			err := d.check(index, size, source)
	//			if err != nil {
	//				log.Error(err)
	//				return
	//			}
	//		}(i*t, size, ctx)
	//	}
	//	err := d.check(n*t, (d.BlockSize - n*t)*BLOCKSIZE + d.SeekSize, source)
	//	if err != nil {
	//		return err
	//	}
	//}
	if d.BlockSize < readBlockCount{
		err := d.check(0, d.BlockSize*BLOCKSIZE + d.SeekSize)
		if err != nil {
			return err
		}
	}else{
		err := d.check(d.BlockSize - readBlockCount, readBlockCount*BLOCKSIZE + d.SeekSize)
		if err != nil {
			return err
		}
	}
	if float64(d.BlockSize * BLOCKSIZE + d.SeekSize) >= (d.Size - 10*MB){
		return fmt.Errorf("disk read is over!!!" )
	}
	writeSize := float64(d.BlockSize * BLOCKSIZE + d.SeekSize)/ MB
	log.Info("****************************************************************")
	log.Infof("%v: total checkout %v MB is same, block size is %v ", d.DiskPath, writeSize, d.BlockSize)
	log.Info("****************************************************************")
	return nil
}

func (d *DiskSizeInfo)read(ctx context.Context, timer *timerTask.TimerConfig)  {
LOOP:
	for{
		if float64(d.BlockSize * BLOCKSIZE + d.SeekSize) >= (d.Size - 10*MB){
			continue
		}

		err := d.CheckReadDisk(ctx)
		if err != nil {
			if strings.Contains(err.Error(), "disk read is over!!!"){
				log.Info(err)
				timer.Stop()
				timer.Waiter.Done()
				return
			}
			log.Warn(err)
		}
		select {
		case <-ctx.Done(): // 等待上级通知
			break LOOP
		default:
		}
	}
}

func (d *DiskSizeInfo)write(ctx context.Context, timer *timerTask.TimerConfig)  {
LOOP:
	for{
		err:= d.DiskWriteByFile()
		if err != nil {
			if strings.Contains(err.Error(), "disk write is over!!!"){
				log.Info(err)
				timer.Waiter.Done()
				return
			}
			log.Warnf("%v :write disk is error: %v, write in to disk is %v byte\n", d.DiskPath, err, d.SeekSize)
		}
		select {
		case <-ctx.Done(): // 等待上级通知
			break LOOP
		default:
		}
	}
}

// 循环通过数据集进项校验:
func (d *DiskSizeInfo)Run() error {
	if !d.DiskStatus(){
		log.Error("未发现磁盘")
        return fmt.Errorf("未发现磁盘")
    }
	log.Infof("%v: disk size: %v\n", d.DiskPath, d.Size)
	// 上电
	d.PowerIn()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	t1 := TimerTime * time.Minute + 50 * time.Second
	log.Infof("开始定时%v分钟断电", TimerTime)
	tasks := []func(){d.PowerOut}
	timer := timerTask.NewTimerTask(t1, tasks)
	timer.Waiter.Add(2)
	timer.Start()
	go d.write(ctx, timer)
	go d.read(ctx, timer)
	timer.Waiter.Wait()
	return nil
}

// get disk status
func (d *DiskSizeInfo)DiskStatus() bool {
	cmd := exec.Command("lsblk")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Println(err.Error())
		return false
	}
	strs := out.String()
	data := strings.Split(d.DiskPath, "/")
	if len(data)<= 0{
		if !strings.Contains(strs, d.DiskPath){
			return false
		}
	}else{
		if !strings.Contains(strs, data[len(data)-1]){
			return false
		}
	}
	destination, err :=  directio.OpenFile(d.DiskPath, os.O_RDONLY, 0666)
	defer destination.Close()
	if err != nil {
		//log.Errorf("open %v disk error:%v ",d.DiskPath, err)
		//destination.Close()
		return false
	}
	//defer destination.Close()
	return true
}

// 断电操作
func (d *DiskSizeInfo)PowerOut()  {
	if !d.DiskStatus(){
		if !d.IsPower{
			fmt.Println("磁盘已断电")
			return
		}
	}
	log.Info("磁盘开始断电！！！")
	Str := fmt.Sprintf("i2cset -f -y 2 0x20 0x3 0xCf")
	cmd := exec.Command("/bin/sh", "-c", Str)
	err := cmd.Run()
	if err != nil {
		//log.Info("断电失败")
		log.Panic("断电失败")
		//return
	}
	//log.Info(d.DiskPath,": 延时40秒下电")
	time.Sleep(40*time.Second)
	if d.DiskStatus(){
		d.powerOutNum++
		if d.powerOutNum > 10{
			d.powerOutNum = 0
			log.Panic("断电失败, 重复断电10次，怀疑磁盘损坏")
		}
		d.PowerOut()
	}
	d.powerOutNum = 0
	d.IsPower = false
	log.Info("磁盘断电成功！！！")
	d.OutPowerNum ++
	log.Infof("%v:磁盘断电次数为：%v", d.DiskPath, d.OutPowerNum)
}

// 上电操作
func (d *DiskSizeInfo)PowerIn(){
	d.diskMutex.Lock()
	defer d.diskMutex.Unlock()
	if d.DiskStatus(){
		if d.IsPower {
			fmt.Println("磁盘已上电")
			return
		}
	}
look:
	Str := fmt.Sprintf("i2cset -f -y 2 0x20 0x3 0xff")
	cmd := exec.Command( "/bin/sh", "-c", Str)
	err := cmd.Run()
	if err != nil {
		//log.Info("磁盘上电失败")
		log.Panic(" 磁盘上电失败")
	}
	//log.Info(d.DiskPath,": 延时10秒上电")
	time.Sleep(10*time.Second)
	if !d.DiskStatus() {
		d.powerInNum++
		if d.powerInNum > 10{
			d.powerInNum = 0
			log.Panic("上电失败, 重复上电10次，怀疑磁盘损坏")
		}
		goto look
	}
	d.powerInNum = 0
	d.IsPower = true
	log.Info(d.DiskPath,": 磁盘上电成功！！！")
}

// 初始化磁盘校验
func InitDiskTest(diskPath string)(*DiskSizeInfo, error) {
	Size := DiskSize(diskPath)
	if Size <= BLOCKSIZE {
		return nil, fmt.Errorf("small disk data, size is:%v \n", Size)
	}
	disk := &DiskSizeInfo{
		DiskPath:  diskPath,
		Size:      Size,
		SeekSize:  0,
		BlockSize: 0,
		IsPower:   	 false,
		OutPowerNum: 0,
		diskContext: nil,
		diskMutex:   sync.Mutex{},
		powerInNum:     0,
		powerOutNum:    0,
	}
	sourceData()
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