package datetype

type AppConfig struct {
	Name          string `mapstructure:"name"`
	Mode          string `mapstructure:"mode"`
	Version       string `mapstructure:"version"`
	Port          int    `mapstructure:"port"`
	StartTime     string `mapstructure:"start_time"`
	MachineId     int64  `mapstructure:"machine_id"`
	ClientIp      string `mapstructure:"clientip"`
	*LogConfig    `mapstructure:"log"`
	*ServerConfig `mapstructure:"server"`
	*EtcdConfig   `mapstructure:"etcd"`
}

type ServerConfig struct {
	Ip        string `mapstructure:"ip"`
	Port      int    `mapstructure:"port"`
	Heartbeat int    `mapstructure:"heartbeat"`
	HostName  string `mapstructure:"hostname"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
}

type EtcdConfig struct {
	Endpoints   []string `mapstructure:"host"`
	DialTimeout int64    `mapstructure:"dialtiemeout"`
	Username    string   `mapstructure:"username"`
	Password    string   `mapstructure:"password"`
}
