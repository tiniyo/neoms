package managers

import (
	"fmt"
	"github.com/tiniyo/neoms/config"
	"github.com/tiniyo/neoms/helper"
	"github.com/tiniyo/neoms/logger"
)

func postRecordingData(callSid string, evHeader []byte)  {
	dataMap := make(map[string]interface{})
	if err := json.Unmarshal(evHeader, &dataMap); err != nil {
		logger.UuidLog("Err", callSid, fmt.Sprint("post recordings failed - ", err))
		return
	}
	recordingServiceUrl := fmt.Sprintf("%s", config.Config.RecordingService.BaseUrl)
	statusCode, _, err := helper.Post(callSid,dataMap, recordingServiceUrl)
	if err != nil {
		logger.UuidLog("Err", callSid, fmt.Sprint("post recordings failed - ", err))
	} else if statusCode != 200 && statusCode != 201 {
		logger.UuidLog("Err", callSid, fmt.Sprint("post recordings failed - ", statusCode))
	} else {
		logger.UuidLog("Info", callSid, "recordings success response received")
	}
}

