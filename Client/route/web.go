package route

import (
	"Client/bin"
	"Client/datetype"
	"Client/logger"
	"Client/mode"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func Setup() *gin.Engine {
	r := gin.New()
	r.Use(logger.GinLogger(), logger.GinRecovery(true))
	r.GET("/api/monitor", MonitorHandler)
	r.GET("/demo", DemoSendDate)
	return r
}
func MonitorHandler(c *gin.Context) {
	// 获取内存数据
	memData, err := bin.GetMemoryData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取内存数据出错",
		})
		log.Println("获取内存数据出错：", err)
		return
	}

	// 获取网络数据
	netData, err := bin.GetNetworkData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取网络数据出错",
		})
		log.Println("获取网络数据出错：", err)
		return
	}

	// 获取 CPU 数据
	cpuData, err := bin.GetCPUInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取 CPU 数据出错",
		})
		log.Println("获取 CPU 数据出错：", err)
		return
	}

	// 获取磁盘数据
	diskData, err := bin.GetDiskInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取磁盘数据出错",
		})
		log.Println("获取磁盘数据出错：", err)
		return
	}
	// 转换 memory 数据单位
	memData.Total, _ = mode.ConvertMemory(memData.Total)
	memData.Used, _ = mode.ConvertMemory(memData.Used)
	memData.Free, _ = mode.ConvertMemory(memData.Free)

	// 转换 disk 数据单位
	for i := range diskData {
		diskData[i].Total = mode.ConvertDisk(diskData[i].Total)
		diskData[i].Used = mode.ConvertDisk(diskData[i].Used)
		diskData[i].Free = mode.ConvertDisk(diskData[i].Free)
	}

	// 转换 network 数据单位
	for i := range netData {
		netData[i].BytesSent, _ = mode.ConvertNetwork(netData[i].BytesSent)
		netData[i].BytesRecv, _ = mode.ConvertNetwork(netData[i].BytesRecv)
	}

	// 构造最终返回的数据结构
	monitorData := datetype.MonitorData{
		Memory:  memData,
		Network: netData,
		CPU:     cpuData,
		Disk:    diskData,
	}

	// 返回 JSON 数据
	c.JSON(http.StatusOK, monitorData)
}
func DemoSendDate(c *gin.Context) {
	// 获取内存数据
	data, err := bin.SendData()
	if err != nil {
		return
	}

	// 返回 JSON 数据
	c.JSON(http.StatusOK, data)
}
