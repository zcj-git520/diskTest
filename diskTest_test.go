package diskTest

import (
	"fmt"
	"testing"
)

func TestDisk( *testing.T)  {
	diskPath := "/dev/sdf"
	srcMap :=  make(map[string]string)
	srcMap["/home/zcj-ubuntu/go_word/demo_0.txt"] = "zhaochengji"
	srcMap["/home/zcj-ubuntu/go_word/demo_1.txt"] = "checkout ser mask"
	srcMap["/home/zcj-ubuntu/go_word/demo_2.txt"] = "./unix.socket"
	size := MB * 100
	seekSzie := size/10
	disk, err := InitDiskTest(diskPath, srcMap, size, seekSzie, 2*BLOCKSIZE)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = disk.Run()
	if err != nil {
		fmt.Println(err)
	}
}
