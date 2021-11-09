package managers

import (
	"fmt"
	"github.com/tiniyo/neoms/logger"
)

type ConfManagerInterface interface {
	CreateConf(from string, to string)
	GetConf(confId string)
	DeleteConf(confId string)
	CreateConference(uuid string, name string, authId string) string
}

type ConfManager struct{}

func (cm ConfManager) CreateConf(from string, to string) {

}

func (cm ConfManager) GetConf(confId string) {

}

func (cm ConfManager) DeleteConf(confId string) {

}

func (cm ConfManager) CreateConference(uuid string, name string, authId string) string {
	logger.Logger.Debug("Creating Conference for " + uuid + "with name " + name)
	confName := fmt.Sprintf("%s-%s@tiniyo", authId, name)
	return confName
}
