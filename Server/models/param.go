package models

// 定义请求的参数结构体
// ErrorResponse 表示 API 的错误响应

type ErrorResponse struct {
	Data string `json:"data"`
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// 用户注册参数

type ParamSignUp struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	RePassword string `json:"re_password" binding:"required,eqfield=Password"`
	Email      string `json:"email"`
}

type ClientParame struct {
	Hostid             int64  `json:"hostid"`
	Hostname           string `json:"hostname"`
	OptionTime         string `json:"optiontime"`
	OptionNote         string `json:"optionnote"`
	OptionIp           string `json:"optionip"`
	OpitonParame       string `json:"opitonparame"`
	OptionParameCpu    string `json:"optionparamecpu"`
	OptionParameMemory string `json:"optionparamememory"`
	OptionParameDisk   string `json:"optionparamedisk"`
	OptionParameUns    string `json:"optionparameuns"`
	OptionParameDns    string `json:"optionparamedns"`
}

type LoginUserinfo struct {
	UserName string `json:"userName" binding:"required"`
	UserCode int    `json:"UserCode"`
}

type NotiAPI struct {
	WorkApiUrl *string `json:"workapiurl"`
	DingApiUrl *string `json:"dingapiurl"`
	DingAtuser *string `json:"dingatuser"`
	WorkAtuser *string `json:"workatuser"`
	Text       string  `json:"content"`
}

type Filelog struct {
	FileName   string `json:"filename"`
	FileId     int64  `json:"fileid"`
	Uploadtime string `json:"uploadtime"`
	FileSize   int64  `json:"filesize"`
	FileDir    string `json:"filedir"`
}

type FileOption struct {
	FileId     int64  `json:"fileid"`
	FileName   string `json:"filename"`
	FileInfo   string `json:"fileinfo"`
	FileOption string `json:"fileoption"`
	OptionTime string `json:"optiontime"`
}
