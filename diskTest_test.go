package diskTest

import (
	"testing"
)

func TestDisk( t *testing.T)  {
	diskPath := "/dev/sdf"
	disk, err := InitDiskTest(diskPath)
	if err != nil {
		t.Errorf("%v 磁盘初始化失败, 失败原因为:%v \n", diskPath, err)
		return
	}
	if !disk.DiskStatus(){
		t.Errorf("%v 磁盘未开启", diskPath)
		return
	}
	err = disk.Run()
	if err != nil {
		t.Errorf(err.Error())
		return
	}

}
