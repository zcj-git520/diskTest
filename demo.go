package main

import (
	"diskTest"
	"flag"
	"fmt"
	"os"
)
var (
	h      bool
	r      int
	diskPath string

)

var log = diskTest.LogNew.Logger

func init()  {
	flag.BoolVar(&h, "h", false, "help")
	flag.StringVar(&diskPath, "c", "/dev/sdd", "diskPath")
	flag.IntVar(&r, "r", 1, "Number of disk read and write rounds. If the value is less than or equal" +
		" to 0, the value is infinite")
	flag.Usage = usage
}

func usage() {
	_, _ = fmt.Fprintf(os.Stderr, `diskTest version: diskTest/1.0.0
Usage: ./diskTest [-h help] [-t ][-c configuration disk path]

Options:
`)
	flag.PrintDefaults()
}


func run() error {
	disk, err := diskTest.InitDiskTest(diskPath)
	if err != nil {
		log.Warnf("%v 磁盘初始化失败, 失败原因为:%v \n", diskPath, err)
		return err
	}
	if !disk.DiskStatus(){
		log.Infof("%v 磁盘未开启", diskPath)
		return  fmt.Errorf("%v 磁盘未开启", diskPath)
	}
	err = disk.Run()
	if err != nil {
		log.Warn(err)
		return err
	}
	return nil
}

func main() {
	flag.Parse()
	if h{
		flag.Usage()
		return
	}
	if r <= 0{
		i := 1
		for{
			err := run()
			if err != nil {
				log.Error(err)
			}
			log.Infof("磁盘读写%v轮!!!!!!", i)
			i++
		}
	}else{
		for i:= 0; i < r;i++{
			err := run()
			if err != nil {
				log.Error(err)
			}
			log.Infof("磁盘读写%v轮!!!!!!", i+1)
		}
	}
}

