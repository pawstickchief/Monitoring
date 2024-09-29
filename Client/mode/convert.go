package mode

import (
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"math"
)

func RoundToTwoDecimal(value float64) float64 {
	return math.Round(value*100) / 100
}

func ConvertMemory(bytes float64) (float64, string) {
	if bytes >= 1024*1024*1024 {
		return RoundToTwoDecimal(bytes / (1024 * 1024 * 1024)), "GB"
	} else {
		return RoundToTwoDecimal(bytes / (1024 * 1024)), "MB"
	}
}

func ConvertDisk(bytes float64) float64 {
	return RoundToTwoDecimal(bytes / (1024 * 1024 * 1024))
}

func ConvertNetwork(bytes float64) (float64, string) {
	if bytes >= 1000*1024 {
		return RoundToTwoDecimal(bytes / (1024 * 1024)), "MB"
	} else {
		return RoundToTwoDecimal(bytes / 1024), "kB"
	}
}
func ConvertGBKToUTF8(gbkData []byte) (string, error) {
	utf8Data, _, err := transform.String(simplifiedchinese.GBK.NewDecoder(), string(gbkData))
	if err != nil {
		return "", err
	}
	return utf8Data, nil
}
