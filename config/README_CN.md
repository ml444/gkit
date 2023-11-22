# config

config模块用于从命令行、环境变量、配置文件中读取配置信息，配置文件支持多种配置格式，如json、yaml、ini、toml等。
并且在设置读取命令行和环境变量时可以配置其默认值（通过结构体的Tag来设置其默认值），当配置文件中没有配置时，会使用默认值。

其读取配置文件的优先级为：`命令行 > 环境变量 > 配置文件 > 命令行默认值(Tag非零默认值) > 环境变量默认值(Tag非零默认值) > 字段零值`
。
如果命令行设定的值为零值，但是环境变量设置的值不为零值，则使用环境变量的值。
零值的优先级虽然最低但提供了最低的保障（防止nil)，即如果配置文件中没有配置，且命令行和环境变量中也没有配置，则使用字段的零值。
默认情况下，config设置的结构体必须时可导出的，非导出的字段将会被忽略，且所有未被设置默认值的字段(包括嵌套结构体)。
都会初始化为零值，防止nil的出现。

## 快速使用
自定义一个config结构体，然后调用`config.InitConfig`方法，传入配置结构体的指针，即可返回一个已经从命令行、环境变量以及配置文件中设置值的结构体指针。

```go
package main

import (
	"time"
	"github.com/ml444/gkit/config"
)

type CustomCfg struct {
	Str      string        `env:"name=GKIT_CONFIG_STR;default=hello"`
	Int      int           `env:"name=GKIT_CONFIG_INT;default=1"`
	Duration time.Duration `env:"name=GKIT_CONFIG_DURATION;default=1s"`
	DBCfg    DBConfig      `env:"name=GKIT_CONFIG_DB"`
}
type DBConfig struct {
	Host     string `env:"name=GKIT_CONFIG_DB_HOST;default=localhost"`
	Username string `env:"name=GKIT_CONFIG_DB_USERNAME;default=root"`
	Port     int    `env:"name=GKIT_CONFIG_DB_PORT;default=3306"`
	Password string `env:"name=GKIT_CONFIG_DB_PASSWORD;default=root"`
}

func main() {
	cfg := &CustomCfg{}
	var _, _ = config.InitConfig(cfg)
	if cfg.DBCfg.Port == 3306 {
        panic("error")
    }
	
	// OR load config from file
	var _, _ = config.InitConfig(cfg, config.WithFilePath("/your/path/config.json"))
	if cfg.DBCfg.Port == 3306 {
        panic("error")
    }
}
```

## 注意

`map[string]interface{}`类型的字段设置默认值时，无法判断`interface{}`的类型，所以其值为字符串类型。

```go
package main

import (
	"os"
	"github.com/ml444/gkit/config"
)

type Cfg struct {
	Map map[string]interface{} `env:"name=GKIT_CONFIG_MAP;default=a:1,b:2"`
}

var _, _ = config.InitConfig(&Cfg{})

// return &Cfg{Map: map[string]interface{}{"a": "11", "b": "22"}}

// OR
func main() {
	_ = os.Setenv("GKIT_CONFIG_MAP", "a:11,b:22")
}

// return &Cfg{Map: map[string]interface{}{"a": "11", "b": "22"}}

```
