package setting

import (
	"Client/datetype"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var Conf = new(datetype.AppConfig)

func Init(configfile string) (err error) {
	viper.SetConfigFile(configfile)
	//指定配置文件
	err = viper.ReadInConfig()
	if err != nil {
		fmt.Printf("Profile read failed, please specify the configuration file:%v\n", err)
		return
	}
	if err := viper.Unmarshal(Conf); err != nil {
		fmt.Printf("viper.Unmarshal failed, err:%v\n", err)
	}
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		fmt.Println("配置文件修改...")
		if err := viper.Unmarshal(Conf); err != nil {
			fmt.Printf("viper.Unmarshal failed, err:%v\n", err)
		}
	})
	return
}
