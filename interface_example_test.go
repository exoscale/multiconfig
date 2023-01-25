package multiconfig_test

import (
	"encoding/json"
	"fmt"

	"github.com/exoscale/multiconfig"
)

// Config is an example application config
type Config struct {
	MyFlag bool
	Mode   string `default:"fast"`
	ServerConfig
}

// ServerConfig is generic server configuration that
// supplies a default for the Port
type ServerConfig struct {
	Endpoint string `default:"localhost"`
	Port     uint   `default:"8080"`
}

// ApplyDefaults overwrites the default port supplied
// by the embedded ServerConfig
func (c *Config) ApplyDefaults() {
	c.ServerConfig.Port = 9090
}

func ExampleInterfaceLoader() {
	// We apply the defaults supplied by tags before we apply the defaults supplied by the interface
	loader := multiconfig.MultiLoader(&multiconfig.TagLoader{}, &multiconfig.InterfaceLoader{})
	conf := &Config{}

	if err := loader.Load(conf); err != nil {
		fmt.Println("Failed to load config: ", err)
	}
	output, err := json.Marshal(conf)
	if err != nil {
		fmt.Println("Failed to marshal configuration: ", err)
	}
	fmt.Printf("%s\n", output)

	// Output:
	// {"MyFlag":false,"Mode":"fast","Endpoint":"localhost","Port":9090}
}
