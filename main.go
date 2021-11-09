package main

import (
	"github.com/tiniyo/neoms/config"
	"github.com/tiniyo/neoms/logger"
	"github.com/tiniyo/neoms/server"
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
