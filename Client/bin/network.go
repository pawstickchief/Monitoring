package bin

import (
	"awesomeProject/datetype"
	"github.com/shirou/gopsutil/v3/net"
)

func GetNetworkData() ([]datetype.NetworkData, error) {
	netIOs, err := net.IOCounters(true)
	if err != nil {
		return nil, err
	}

	var netData []datetype.NetworkData
	for _, netIO := range netIOs {
		// 跳过环回接口和无活动的接口
		if netIO.BytesRecv == 0 && netIO.BytesSent == 0 {
			continue
		}
		if netIO.Name == "lo" {
			continue
		}
		data := datetype.NetworkData{
			Name:        netIO.Name,
			BytesSent:   float64(netIO.BytesSent),
			BytesRecv:   float64(netIO.BytesRecv),
			PacketsSent: float64(netIO.PacketsSent),
			PacketsRecv: float64(netIO.PacketsRecv),
		}
		netData = append(netData, data)
	}

	return netData, nil
}
