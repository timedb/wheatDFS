package serverTorch

import (
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"math/rand"
	"os"
	"runtime"
	"time"
)

//初始化随机数
func init() {
	rand.Seed(time.Now().Unix())
}

// GetCpuPercent 获取CPU占用率
func GetCpuPercent() float64 {
	percent, _ := cpu.Percent(time.Second, false)
	return percent[0]
}

// GetMemPercent 获取内存占用率
func GetMemPercent() float64 {
	memInfo, _ := mem.VirtualMemory()
	return memInfo.UsedPercent
}

// GetDiskPercent 获取磁盘占用率
func GetDiskPercent() float64 {
	parts, _ := disk.Partitions(true)
	diskInfo, _ := disk.Usage(parts[0].Mountpoint)
	return diskInfo.UsedPercent
}

// GetServerWeight 得到权重
func GetServerWeight() int {
	CoreNum := runtime.NumCPU() // cpu核数获取
	percent, _ := cpu.Percent(time.Second, false)
	UsePercent := (100 - percent[0]) / 100 // cpu剩余
	memInfo, _ := mem.VirtualMemory()
	memTotal := memInfo.Total / 1024 / 1024 / 1024 // 内存
	memLeft := memInfo.Free / 1024 / 1024 / 1024   // 内存剩余量
	point := float64(CoreNum)*20*UsePercent + float64(memTotal)*float64(memLeft)*40
	result := int(point)

	result += rand.Intn(result / 5) //附加小权重

	return result / 10

}

// CreteMkdir 创建group
func CreteMkdir(path string) {

	if path[len(path)-1] != '/' {
		path = path + "/"
	}

	_path := fmt.Sprintf(path)
	_ = os.Mkdir(_path, os.ModePerm)
	var initDir = [16]string{"0", "1", "2", "3", "4", "5", "6", "7",
		"8", "9", "a", "b", "c", "d", "e", "f"}
	var pa []string
	for _, v1 := range initDir {
		for _, v2 := range initDir {
			pa = append(pa, v1+v2)
		}
	}
	for _, v1 := range pa {
		if _, err := os.Stat(_path + v1); err == nil {
			continue
		}
		_ = os.Mkdir(_path+v1, os.ModePerm)
		for _, v2 := range pa {
			_ = os.Mkdir(_path+v1+"/"+v2, os.ModePerm)
		}
	}
}

// JoinPath 拼接地址
// JoinPath("sto/", "pdo", "fq")-> sto/pdo/fq
func JoinPath(paths ...string) string {
	result := ""

	for _, val := range paths {
		if val[len(val)-1] == '/' {
			val = val[:len(val)-1]
		}

		result += val + "/"

	}

	return result[:len(result)-1]
}
