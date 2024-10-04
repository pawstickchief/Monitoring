package datetype

type AppConfig struct {
	Name          string `mapstructure:"name"`
	Mode          string `mapstructure:"mode"`
	Version       string `mapstructure:"version"`
	Port          int    `mapstructure:"port"`
	StartTime     string `mapstructure:"start_time"`
	MachineId     int64  `mapstructure:"machine_id"`
	ClientIp      string `mapstructure:"clientip"`
	WorkDir       string `mapstructure:"workdir"`
	*LogConfig    `mapstructure:"log"`
	*ServerConfig `mapstructure:"server"`
	*EtcdConfig   `mapstructure:"akile"`
	*WebSocket    `mapstructure:"websocket"`
}

type ServerConfig struct {
	Ip          string `mapstructure:"ip"`
	Port        int    `mapstructure:"port"`
	Heartbeat   int    `mapstructure:"heartbeat"`
	HostName    string `mapstructure:"hostname"`
	DownloadApi string `mapstructure:"download_api"`
	UploadApi   string `mapstructure:"upload_api"`
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

type WebSocket struct {
	ReconnectAttempts         int `mapstructure:"reconnect_delay"`
	ReconnectDelay            int `mapstructure:"reconnect_delay"`
	RetryAfterFailure         int `mapstructure:"reconnect_delay"`
	ReconnectInterval         int `mapstructure:"reconnect_interval"`
	MaxTotalReconnectAttempts int `mapstructure:"max_total_reconnect_attempts"`
}
