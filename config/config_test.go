package config

import (
	"os"
	"reflect"
	"testing"
)

type testConfig struct {
	Anonymous
	disable       bool
	Debug         bool                   `env:"name=GKIT_CONFIG_DEBUG;default=false" json:"debug,omitempty" flag:"name=debug;default=true;usage='provide additional information or enable certain debugging'"`
	String        string                 `env:"name=GKIT_CONFIG_String;default=this_is_default"`
	UintPtr       *uint                  `env:"name=GKIT_CONFIG_UintPtr;default=123"`
	Uint          uint                   `env:"name=GKIT_CONFIG_Uint;default=123"`
	Uint8         uint8                  `env:"name=GKIT_CONFIG_Uint8;default=8"`
	Uint16        uint16                 `env:"name=GKIT_CONFIG_Uint16;default=16"`
	Uint32        uint32                 `env:"name=GKIT_CONFIG_Uint32;default=32"`
	Uint64        uint64                 `env:"name=GKIT_CONFIG_Uint64;default=64"`
	Int           int                    `env:"name=GKIT_CONFIG_Int;default=456"`
	Int8          int8                   `env:"name=GKIT_CONFIG_Int8;default=8"`
	Int16         int16                  `env:"name=GKIT_CONFIG_Int16;default=16"`
	Int32         int32                  `env:"name=GKIT_CONFIG_Int32;default=32"`
	Int64         int64                  `env:"name=GKIT_CONFIG_Int64;default=64"`
	Float32       float32                `env:"name=GKIT_CONFIG_Float32;default=3.14"`
	Float64       float64                `env:"name=GKIT_CONFIG_Float64;default=3.14"`
	GroupIDList   []uint32               `env:"name=GKIT_CONFIG_GROUP_ID_LIST;default=1,2"`
	GroupNameList []string               `env:"name=GKIT_CONFIG_GROUP_NAME_LIST;default=group1,group2"`
	Map           map[string]interface{} `env:"name=GKIT_CONFIG_MAP;default=a:1,b:2"`
	DBCfg         dbCfg
	DBCfgPtr      *dbCfg
	RedisCfg      struct {
		DB   int
		User string
		PWD  string
	}
}
type Anonymous struct {
	Account    []*string
	Password   []string
	permission uint64
}
type dbCfg struct {
	URI             string `json:"uri"`
	ConnMaxLifeTime int    `json:"conn_max_life_time" env:"name=CONN_MAX_LIFE_TIME"`
	ConnMaxIdleTime int    `json:"conn_max_idle_time"`
	MaxOpenConns    int    `json:"max_open_conns"`
	MaxIdleConns    int    `json:"max_idle_conns"`
	Acc             *Anonymous
}

func TestNewConfig(t *testing.T) {
	//cfg0 := InitConfig(testConfig{})
	//t.Log(cfg0)
	os.Setenv("CONN_MAX_LIFE_TIME", "1234567890")
	s1 := "fo"
	s2 := "ba"
	cfg := &testConfig{
		GroupNameList: []string{"${HOME}", "group2"},
		DBCfgPtr: &dbCfg{
			URI:             "This is a ${HOME} directory and it belongs to ${USER}.",
			ConnMaxLifeTime: 1,
			ConnMaxIdleTime: 2,
			MaxOpenConns:    3,
			MaxIdleConns:    4,
			Acc: &Anonymous{
				Account:    []*string{&s1, &s2},
				Password:   []string{s1, s2},
				permission: 0,
			},
		}}
	cfgDefault, err := InitConfig(cfg)
	if err != nil {
		t.Error(err)
	}
	if cfg.Debug != true {
		t.Error("the flag default is not set correctly")
	}
	if cfg.String != "this_is_default" {
		t.Error("the env default is not set correctly")
	}
	if cfg.Uint != 123 {
		t.Error("the env default is not set correctly")
	}
	if reflect.DeepEqual(cfg.Map, map[string]interface{}{"a": 1, "b": 2}) {
		t.Error("the env default is not set correctly")
	}
	if cfg.GroupNameList[0] == "${HOME}" {
		t.Error("the env default is not set correctly")
	} else {
		t.Log(cfg.GroupNameList)
	}
	_ = cfgDefault.SetAndChangeEnv("String", "test")
	_ = cfgDefault.SetAndChangeEnv("Uint", "456")
	_ = cfgDefault.SetAndChangeEnv("UintPtr", "456")
	_ = cfgDefault.SetAndChangeEnv("Map", "a:11,b:22")
	if cfg.String != "test" {
		t.Error("don't get the value of env")
	}
	if cfg.Uint != 456 {
		t.Error("don't get the value of env")
	}
	if cfg.UintPtr != nil && *cfg.UintPtr != 456 {
		t.Error("don't get the value of env")
	}
	if !reflect.DeepEqual(cfg.Map, map[string]interface{}{"a": "11", "b": "22"}) {
		t.Error("don't get the value of env")
	}
	t.Log(cfg.DBCfgPtr.URI)
	if cfg.DBCfgPtr.URI == "This is a ${HOME} directory and it belongs to ${USER}." {
		t.Error("ReplaceEnvVariables failed")
	}
	err = cfgDefault.Set("DBCfgPtr__URI", "mysql://root:123456@localhost:3306/test?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		t.Error(err)
	}
	if cfg.DBCfgPtr.URI != "mysql://root:123456@localhost:3306/test?charset=utf8mb4&parseTime=True&loc=Local" {
		t.Error("setIntoStruct failed")
	}
}

type Config1 struct {
	Debug     bool   `yaml:"debug" env:"name=DEBUG" flag:"name=debug;default=true;usage=provide additional information or enable certain debugging"`
	DeployEnv string `yaml:"deploy_env"`

	PProfCfg        *PProfCfg         `yaml:"pprof_cfg"`
	HTTPCfg         *HTTPCfg          `yaml:"http_cfg"`
	GrpcCfg         *GrpcCfg          `yaml:"grpc_cfg"`
	DBCfg           *DBCfg            `yaml:"db_cfg"`
	JobCfgList      []*ScheduleJobCfg `yaml:"job_cfg_list"`
	AccessKeyID     string            `yaml:"access_key_id" env:"name=ALIBABA_CLOUD_ACCESS_KEY_ID"`
	AccessKeySecret string            `yaml:"access_key_secret" env:"name=ALIBABA_CLOUD_ACCESS_KEY_SECRET"`
	MsgCfg          *MsgCfg           `yaml:"msg_cfg"`
	JWTCfg          *JWTCfg           `yaml:"jwt_cfg"`
}
type PProfCfg struct {
	EnablePprof bool   `yaml:"enable_pprof" env:"name=ENABLE_PPROF" flag:"name=enable_pprof;default=false;usage='enable APIs for pprof'"`
	PprofAddr   string `yaml:"pprof_addr" env:"name=PPROF_ADDR" flag:"name=pprof_addr;default=false;usage='pprof listen address'"`
}

type HTTPCfg struct {
	EnableHTTP           bool   `yaml:"enable_http" env:"name=ENABLE_HTTP" flag:"name=enable_http;default=false;usage='enable APIs for http'"`
	HTTPAddr             string `yaml:"http_addr" env:"name=HTTP_ADDR" flag:"name=http_addr;default=false;usage='HTTP listen address'"`
	Timeout              int    `yaml:"timeout"`
	TLSCertFile          string `yaml:"tls_cert_file"`
	TLSKeyFile           string `yaml:"tls_key_file"`
	RouterPathPrefix     string `yaml:"router_path_prefix"`
	RouterStrictSlash    bool   `yaml:"router_strict_slash"`
	RouterSkipClean      bool   `yaml:"router_skip_clean"`
	RouterUseEncodedPath bool   `yaml:"router_use_encoded_path"`
}
type GrpcCfg struct {
	EnableGRPC   bool   `yaml:"enable_grpc" env:"name=ENABLE_GRPC" flag:"name=enable_grpc;default=false;usage=enable APIs for grpc"`
	EnableHealth bool   `yaml:"enable_health" env:"name=ENABLE_GRPC_HEALTH" flag:"name=enable_health;default=false;usage='enable APIs for grpc health'"`
	GRPCAddr     string `yaml:"grpc_addr" env:"name=GRPC_ADDR" flag:"name=grpc_addr;default=false;usage='gRPC listen address'"`
	TLSCertFile  string `yaml:"tls_cert_file"`
	TLSKeyFile   string `yaml:"tls_key_file"`
}
type DBCfg struct {
	DBURI           string `yaml:"db_uri" env:"name=SERVICE_DB_URI"`
	ConnMaxLifeTime int    `yaml:"connmaxlifetime"` // 连接池中每个连接的最大生存时间，单位秒。
	ConnMaxIdleTime int    `yaml:"connmaxidletime"` // 连接池中每个连接的最大空闲时间，单位秒。
	MaxOpenConns    int    `yaml:"maxopenconns"`    // 连接池中允许同时打开的最大连接数
	MaxIdleConns    int    `yaml:"maxidleconns"`    // 连接池中允许存在的最大空闲连接数
}

type ScheduleJobCfg struct {
	JobName  string `yaml:"job_name"`
	Interval int    `yaml:"interval"`
	AtTime   string `yaml:"at_time"`
	CronExp  string `yaml:"cron_exp"`
}

type MsgCfg struct {
	RequestURL     string   `yaml:"agent_url"`
	EnableTestSend bool     `yaml:"enable_test_send"`
	TestGroups     []string `yaml:"test_mango_groups"`
}
type JWTCfg struct {
	Secret string `yaml:"secret"`
}

func TestConfig1(t *testing.T) {
	cfg := Config1{}
	c, err := InitConfig(&cfg)
	if err != nil {
		t.Error(err)
	}
	t.Log(c)
	t.Log(cfg.Debug)
}
