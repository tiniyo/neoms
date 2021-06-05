package main

import (
	"github.com/neoms/config"
	"github.com/neoms/logger"
	"github.com/neoms/server"
)

/*
	WebMediaServer :- Initialize server and configuration
 */
func main() {
	config.InitConfig()
	//helper.InitHttpConnPool()
	logger.InitLogger()
	server.Init()
}
