package config

import (
	"flag"
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
	err := InitConfig(cfg)
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

	t.Log(cfg.DBCfgPtr.URI)
	if cfg.DBCfgPtr.URI == "This is a ${HOME} directory and it belongs to ${USER}." {
		t.Error("ReplaceEnvVariables failed")
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
	err := InitConfig(&cfg)
	if err != nil {
		t.Error(err)
	}
	t.Log(cfg.Debug)
}

func TestWalk(t *testing.T) {
	type args struct {
		c  any
		fn func(k string, v any) error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				c: &testConfig{Debug: true},
				fn: func(k string, v any) error {
					t.Logf("key: %s, value: %v \n", k, v)
					return nil
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Walk(tt.args.c, tt.args.fn); (err != nil) != tt.wantErr {
				t.Errorf("Walk() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type Inner struct {
    Name *string
    Age  *int
}

type TestStruct struct {
    Field1 string
    Field2 *string
    Slice  []string
    Inner  *Inner
}


func strptr(s string) *string { return &s }
func intptr(i int) *int       { return &i }

func TestBuildMap_BasicStruct(t *testing.T) {
	cfg := TestStruct{}

	p := reflect.ValueOf(&cfg).Elem()
	proc := &Processor{m: make(map[string]*Value)}

	if err := proc.buildMap("", p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 验证 Field1 初始化为 ""（zero value）
	v1 := proc.m["Field1"].value.(string)
	if v1 != "" {
		t.Errorf("expected Field1 zero value, got %q", v1)
	}
}

func TestBuildMap_StringPointerInit(t *testing.T) {
	cfg := TestStruct{Field2: nil}

	p := reflect.ValueOf(&cfg).Elem()
	proc := &Processor{m: make(map[string]*Value)}

	if err := proc.buildMap("", p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 指针字段应该初始化，并且在 map 中值为 ""
	v2 := proc.m["Field2"].value.(string)
	if v2 != "" {
		t.Errorf("expected Field2 zero string, got %q", v2)
	}
}

func TestBuildMap_SliceString(t *testing.T) {
	cfg := TestStruct{Slice: []string{"A", ""}}

	p := reflect.ValueOf(&cfg).Elem()
	proc := &Processor{m: make(map[string]*Value)}

	if err := proc.buildMap("", p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	gotSlice, ok := proc.m["Slice"].value.([]string)
	if !ok {
		t.Fatal("expected []string in map")
	}

	if gotSlice[0] != "A" || gotSlice[1] != "" {
		t.Errorf("slice content mismatch: got %v", gotSlice)
	}
}

func TestBuildMap_NestedStruct(t *testing.T) {
	cfg := TestStruct{
		Inner: &Inner{Name: strptr("X")},
	}

	p := reflect.ValueOf(&cfg).Elem()
	proc := &Processor{m: make(map[string]*Value)}

	if err := proc.buildMap("", p); err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	// Nested 字段 Field: Inner.Name
	val, ok := proc.m["Inner"+delimiter+"Name"].value.(string)
	if !ok {
		t.Errorf("expected string for nested field, got %T", proc.m["Inner"+delimiter+"Name"].value)
	}
	if val != "X" {
		t.Errorf("expected X, got %q", val)
	}
}

type priorityConfig struct {
	Debug bool `env:"name=GKIT_PRIORITY_DEBUG;default=true" flag:"name=debug;default=true;usage=debug flag"`
}

func TestInitConfig_CommandLineZeroOverridesEnv(t *testing.T) {
	t.Setenv("GKIT_PRIORITY_DEBUG", "true")
	cfg := &priorityConfig{}
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	err := InitConfig(
		cfg,
		WithFlagSet(fs),
		WithArgs([]string{"--debug=false"}),
	)
	if err != nil {
		t.Fatalf("InitConfig failed: %v", err)
	}
	if cfg.Debug {
		t.Fatalf("expected command line value false to win over env, got true")
	}
}

type envPrefixConfig struct {
	Port int
}

func TestInitConfig_WithEnvKeyPrefix(t *testing.T) {
	t.Setenv("APP_PORT", "9090")
	cfg := &envPrefixConfig{}
	err := InitConfig(cfg, WithEnvKeyPrefix("APP_"))
	if err != nil {
		t.Fatalf("InitConfig failed: %v", err)
	}
	if cfg.Port != 9090 {
		t.Fatalf("expected Port=9090, got %d", cfg.Port)
	}
}

type envTagEqualConfig struct {
	Token string `env:"name=GKIT_ENV_TOKEN;default=a=b=c"`
}

func TestInitConfig_EnvDefaultSupportsEquals(t *testing.T) {
	cfg := &envTagEqualConfig{}
	err := InitConfig(cfg)
	if err != nil {
		t.Fatalf("InitConfig failed: %v", err)
	}
	if cfg.Token != "a=b=c" {
		t.Fatalf("expected token default to preserve '=', got %q", cfg.Token)
	}
}

func TestInitConfig_FileFlagIgnoreUnknownArgs(t *testing.T) {
	cfg := &envPrefixConfig{}
	err := InitConfig(
		cfg,
		WithFileFlag("config"),
		WithArgs([]string{"--unknown=1"}),
	)
	if err != nil {
		t.Fatalf("expected unknown args ignored by file flag parser, got %v", err)
	}
}

type sameFlagConfigA struct {
	Name string `flag:"name=name;default=alice;usage=name flag"`
}

type sameFlagConfigB struct {
	Name string `flag:"name=name;default=bob;usage=name flag"`
}

func TestInitConfig_DuplicateFlagNameAcrossProcessors(t *testing.T) {
	cfgA := &sameFlagConfigA{}
	cfgB := &sameFlagConfigB{}

	err := InitConfig(
		cfgA,
		WithFlagSet(flag.NewFlagSet("a", flag.ContinueOnError)),
		WithArgs([]string{"--name=tom"}),
	)
	if err != nil {
		t.Fatalf("InitConfig cfgA failed: %v", err)
	}

	err = InitConfig(
		cfgB,
		WithFlagSet(flag.NewFlagSet("b", flag.ContinueOnError)),
		WithArgs([]string{"--name=jerry"}),
	)
	if err != nil {
		t.Fatalf("InitConfig cfgB failed: %v", err)
	}

	if cfgA.Name != "tom" || cfgB.Name != "jerry" {
		t.Fatalf("expected independent flag parsing, got cfgA=%q cfgB=%q", cfgA.Name, cfgB.Name)
	}
}

type envShorthandConfig struct {
	Token string `env:"TOKEN; default=abc"`
}

type envExplicitNameConfig struct {
	Token string `env:"name=TOKEN; default=abc"`
}

func TestInitConfig_EnvShorthandName(t *testing.T) {
	t.Setenv("TOKEN", "from-env")
	cfg := &envShorthandConfig{}
	if err := InitConfig(cfg); err != nil {
		t.Fatalf("InitConfig failed: %v", err)
	}
	if cfg.Token != "from-env" {
		t.Fatalf("expected env value from-env, got %q", cfg.Token)
	}
}

func TestInitConfig_EnvShorthandNameDefault(t *testing.T) {
	t.Setenv("TOKEN", "")
	cfg := &envShorthandConfig{}
	if err := InitConfig(cfg); err != nil {
		t.Fatalf("InitConfig failed: %v", err)
	}
	if cfg.Token != "abc" {
		t.Fatalf("expected default abc, got %q", cfg.Token)
	}
}

func TestInitConfig_EnvExplicitName(t *testing.T) {
	t.Setenv("TOKEN", "explicit-env")
	cfg := &envExplicitNameConfig{}
	if err := InitConfig(cfg); err != nil {
		t.Fatalf("InitConfig failed: %v", err)
	}
	if cfg.Token != "explicit-env" {
		t.Fatalf("expected env value explicit-env, got %q", cfg.Token)
	}
}

type flagShorthandConfig struct {
	Token string `flag:"token; default=abc; usage=bot token"`
}

type flagExplicitNameConfig struct {
	Token string `flag:"name=token; default=abc; usage=bot token"`
}

func TestInitConfig_FlagShorthandName(t *testing.T) {
	cfg := &flagShorthandConfig{}
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	if err := InitConfig(cfg, WithFlagSet(fs), WithArgs([]string{"--token=from-flag"})); err != nil {
		t.Fatalf("InitConfig failed: %v", err)
	}
	if cfg.Token != "from-flag" {
		t.Fatalf("expected flag value from-flag, got %q", cfg.Token)
	}
}

func TestInitConfig_FlagExplicitName(t *testing.T) {
	cfg := &flagExplicitNameConfig{}
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	if err := InitConfig(cfg, WithFlagSet(fs), WithArgs([]string{"--token=from-flag"})); err != nil {
		t.Fatalf("InitConfig failed: %v", err)
	}
	if cfg.Token != "from-flag" {
		t.Fatalf("expected flag value from-flag, got %q", cfg.Token)
	}
}

func TestParseStructTagOptions(t *testing.T) {
	tests := []struct {
		name    string
		values  []string
		want    tagOptions
		wantErr bool
	}{
		{
			name:   "env shorthand with default",
			values: []string{"TOKEN", "default=abc"},
			want:   tagOptions{name: "TOKEN", defaultStr: "abc"},
		},
		{
			name:   "explicit name with default",
			values: []string{"name=TOKEN", "default=abc"},
			want:   tagOptions{name: "TOKEN", defaultStr: "abc"},
		},
		{
			name:   "flag shorthand with usage",
			values: []string{"token", "default=abc", "usage=bot token"},
			want:   tagOptions{name: "token", defaultStr: "abc", usage: "bot token"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseStructTagOptions(tt.values)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseStructTagOptions() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("parseStructTagOptions() = %+v, want %+v", got, tt.want)
			}
		})
	}
}