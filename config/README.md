# config
The config module is used to read configuration information from the command 
line, environment variables, and configuration files. Configuration files 
support multiple configuration formats, such as json, yaml, ini, toml, etc.
And when setting up and reading the command line and environment variables, 
you can configure their default values (set their default values through the 
Tag of the structure). When there is no configuration in the configuration 
file, the default value will be used.

Its priority for reading configuration files is: 
`Command line > Environment variables > Configuration file > Command line default value (Tag non-zero default value) > Environment variable default value (Tag non-zero default value) > Field zero value`.
If the value set on the command line is zero, but the value set by an 
environment variable is not zero, the value of the environment variable is used.
Although the zero value has the lowest priority, it provides the lowest 
guarantee (preventing nil), that is, if there is no configuration in the 
configuration file, and there is no configuration in the command line and 
environment variables, the zero value of the field is used.
By default, the structure set by config must be exportable, non-exported 
fields will be ignored, and all fields that do not have default values 
set (including nested structures).
Will be initialized to zero value to prevent the occurrence of nil.

## QUICK TO USE
Customize a config structure, then call the `config.Init Config` method 
and pass in the pointer of the configuration structure, which will return 
a structure pointer with values set from the command line, environment 
variables and configuration files.
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

## Notice

When setting a default value for a field of type `map[string]interface{}`, 
the type of `interface{}` cannot be determined, so its value is of string type.

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
