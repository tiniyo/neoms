package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

/* TomlConfig represent the root of configuration */
type TomlConfig struct {
	Title   string
	Owner   Owner
	Server  Server
	Logging Logging
	Fs      Freeswitch
	Redis   Redis
	Heartbeat Heartbeat
	Rating	Rating
	Numbers Numbers
	SipEndpoint SipEndpoint
	Kamgo Kamgo
}

type Heartbeat struct {
	BaseUrl string
	UserName string
	Secret string
}

type Rating struct {
	BaseUrl string
	UserName string
	Secret string
	Region string
}
type Numbers struct {
	BaseUrl string
	UserName string
	Secret string
}
type SipEndpoint struct {
	BaseUrl string
	UserName string
	Secret string
}

type Kamgo struct {
	BaseUrl string
	UserName string
	Secret string
}
/* Owner represent owner of the module*/
type Owner struct {
	Name string
	Org  string `toml:"organization"`
	Bio  string
}

type Server struct {
	Port          string
	QueueLen      int
	ErrorQueueLen int
	GinMode       string
}

type Logging struct {
	Facility string
	Level    string
	Tag      string
	Syslog   string
	Sentry   string
}

type Freeswitch struct {
	FsHost     string
	FsPort     string
	FsPassword string
	FsTimeout  int
}

type Redis struct {
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int
}

var Config TomlConfig

func InitConfig() {
	var err error

	configFile := os.Getenv("WEBFS_CONFIG")
	if len(configFile) == 0 {
		configFile = "/etc/config.toml"
	}

	fmt.Println("Config file is ", configFile)

	if _, err = toml.DecodeFile(configFile, &Config); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(Config)
}
