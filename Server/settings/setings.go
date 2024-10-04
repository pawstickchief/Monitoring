package settings

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var Conf = new(AppConfig)

type AppConfig struct {
	Name         string `mapstructure:"name"`
	Mode         string `mapstructure:"mode"`
	Version      string `mapstructure:"version"`
	Port         int    `mapstructure:"port"`
	StartTime    string `mapstructure:"start_time"`
	MachineId    int64  `mapstructure:"machine_id"`
	ClientUrl    string `mapstructure:"client_url"`
	*LogConfig   `mapstructure:"log"`
	*MySQLConfig `mapstructure:"mysql"`
	*FileConfig  `mapstructure:"file"`
	*WXworkToke  `mapstructure:"WXWork"`
	*EtcdConfig  `mapstructure:"etcd"`
}
type FileConfig struct {
	Filemaxsize int64  `mapstructure:"filemaxsize"`
	Savedir     string `mapstructure:"savedir"`
	Httpurl     string `mapstructure:"httpurl"`
	Httpdir     string `mapstructure:"httpdir"`
}
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
}

type MySQLConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DbName       string `mapstructure:"dbname"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxConns     int    `mapstructure:"max_conns"`
}
type WXworkToke struct {
	ApiToken string `mapstructure:"apitoken"`
}
type EtcdConfig struct {
	Endpoints   []string `mapstructure:"endpoints"`
	DialTimeout int      `mapstructure:"dial_timeout"`
	// 如果需要用户名和密码
	EtcdName string `mapstructure:"etcdname"`
	Password string `mapstructure:"password"`
	// 新增 TLS 相关配置
	CaCert     string `mapstructure:"ca_cert"`
	CertFile   string `mapstructure:"cert_file"`
	KeyFile    string `mapstructure:"key_file"`
	ServerName string `mapstructure:"server_name"`
}

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
