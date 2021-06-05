package managers

import (
	"fmt"
	"github.com/neoms/logger"
)

type ConfManager struct{}

func (cm ConfManager) CreateConf(from string, to string) {

}

func (cm ConfManager) GetConf(confid string) {

}

func (cm ConfManager) DeleteConf(confid string) {

}

func (cm ConfManager) CreateConference(uuid string, name string, authId string) string {
	logger.Logger.Debug("Creating Conference for " + uuid + "with name " +name)
	confName := fmt.Sprintf("%s-%s@tiniyo", authId,name)
	return confName
}