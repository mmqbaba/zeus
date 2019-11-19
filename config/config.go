package config

// Configer 配置初始化器
type Configer interface {
	Init(original []byte) error
	Get() *AppConf
}

// Entry 配置入口
type Entry struct {
	ConfigPath   string            `json:"config_path"`   // 配置路径
	ConfigFormat string            `json:"config_format"` // json/toml/yaml
	EngineType   string            `json:"engine_type"`   // etcd/file
	EndPoints    []string          `json:"endpoints"`
	UserName     string            `json:"username"`
	Password     string            `json:"password"`
	Ext          map[string]string `json:"ext"` // 扩展配置
}

// AppConf 应用的具体配置
type AppConf struct {
	LogConf             LogConf                `json:"log_conf" toml:"log_conf" yaml:"log_conf"`
	Redis               Redis                  `json:"redis"`
	MongoDBSource       map[string]MongoDB     `json:"mongodb_source"`
	MysqlSource         map[string]MysqlDB     `json:"mysql_source"`
	RedisSource         map[string]Redis       `json:"redis_source"`
	EBus                EBus                   `json:"ebus"`
	Ext                 map[string]interface{} `json:"ext"`
	Trace               Trace                  `json:"trace"`
	Debug               DebugSwitch            `json:"debug"`
	Obs                 Obs                    `json:"obs"`
	Broker              Broker                 `json:"broker"`
	CurrentBusIdSpIdMap map[string]string      `json:"current_busid_spid_map,omitempty"`
	GoMicro             GoMicro                `json:"go_micro"`
}

type DebugSwitch struct {
	InvalidCorperRetain int32 `json:"invalid_corper_retain"` // 测试用，>=1-保留无效法人代表 0-不保留
}

type Trace struct {
	ServiceName string  `json:"srvice_name"`
	TraceUrl   string  `json:"trace_url"`
	Rate       float64 `json:"rate"`
	Sampler    string  `json:"sampler"`
	Mod        uint64  `json:"mod"`
	OnlyLogErr bool    `json:"only_log_err"` // true 只记录出错日志
	Enable     bool    `json:"enable"`       // 启用组件
}

type MongoDB struct {
	Name            string `json:"name"`
	Host            string `json:"host"`
	User            string `json:"user"`
	Pwd             string `json:"pwd"`
	MaxPoolSize     uint16 `json:"max_pool_size"`
	MaxConnIdleTime uint32 `json:"max_conn_idletime"` // 单位秒
	Enable          bool   `json:"enable"`            // 启用组件
}

type Redis struct {
	Host               string `json:"host"`
	SentinelHost       string `json:"sentinel_host"`
	SentinelMastername string `json:"sentinel_mastername"`
	Pwd                string `json:"pwd"`
	PoolSize           int    `json:"poolsize"`
	Enable             bool   `json:"enable"` // 启用组件
}

type MysqlDB struct {
	DataSourceName string `json:"datasourcename"`
	MaxIdleConns   int    `json:"maxidleconns"`
	MaxOpenConns   int    `json:"maxopenconns"`
	Enable         bool   `json:"enable"` // 启用组件
}

type EBus struct {
	Hosts     []string               `json:"hosts"`
	Services  map[string]interface{} `json:"services"`
	PaasId    string                 `json:"paas_id"`
	PaasToken string                 `json:"paas_token"`
	SpId      string                 `json:"sp_id"`
	PathMap   map[string]string      `json:"path_map"`
}

type LogConf struct {
	Log                 string `json:"log"` // 输出方式：console/file/kafka
	Level               string `json:"level"`
	Format              string `json:"format"`
	RotationTime        string `json:"rotation_time"`
	LogDir              string `json:"log_dir"`
	DisableReportCaller bool   `json:"disable_report_caller"`
}

type Obs struct {
	Endpoint   string `json:"endpoint"`
	AK         string `json:"ak"`
	SK         string `json:"sk"`
	BucketName string `json:"bucket_name"`
	Location   string `json:"location"` // 路径
}

type Broker struct {
	Hosts           []string     `json:"hosts"`
	Type            string       `json:"type"`
	NeedAuth        bool         `json:"need_auth"`
	User            string       `json:"user"`
	Pwd             string       `json:"pwd"`
	TopicPrefix     string       `json:"topic_prefix"`
	SubscribeTopics []*TopicInfo `json:"subscribe_topics"` // 服务订阅的主题
	EnablePub       bool         `json:"enable_pub"`       // 启用pub
}

type TopicInfo struct {
	Category string `json:"category"` // 类别
	Source   string `json:"source"`   // 服务来源
	Queue    string `json:"queue"`    // 队列/组
}

type GoMicro struct {
	ServiceName         string   `json:"service_name"`
	ServerPort         uint32   `json:"server_port"`
	RegistryPluginType string   `json:"registry_plugin_type"` // etcd/consul
	RegistryAddrs      []string `json:"registry_addrs"`       // etcd/consul address
	RegistryAuthUser   string   `json:"registry_authuser"`
	RegistryAutPwd     string   `json:"registry_authpwd"`
}
