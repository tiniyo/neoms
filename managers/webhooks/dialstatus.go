package webhooks

import (
	"fmt"
	"github.com/neoms/helper"
	"github.com/neoms/logger"
	"github.com/neoms/models"
	"strings"
)

func ProcessStatusCallbackUrl(dataCallRequest models.CallRequest, state string) {
	var err error
	callSid := dataCallRequest.CallSid
	dataMap := make(map[string]interface{})
	if callbackByte, err := json.Marshal(dataCallRequest.Callback); err == nil {
		if err := json.Unmarshal(callbackByte, &dataMap); err != nil {
			logger.UuidLog("Err", callSid, fmt.Sprint("call back map conversion issue - ", err))
		}
	}
	if strings.Contains(dataCallRequest.StatusCallbackEvent, state) &&
		len(dataCallRequest.StatusCallback) > 0 && dataCallRequest.StatusCallbackMethod == "GET" {
		_, _, err = helper.Get(callSid,dataMap, dataCallRequest.StatusCallback)
		if err != nil {
			logger.UuidLog("Err", callSid, fmt.Sprint("failed status callback with error ",
				err, " url ", dataCallRequest.StatusCallback))
		}
	} else if strings.Contains(dataCallRequest.StatusCallbackEvent, state) &&
		len(dataCallRequest.StatusCallback) > 0 {
		_, _, err = helper.Post(callSid,dataMap, dataCallRequest.StatusCallback)
		if err != nil {
			logger.UuidLog("Err", callSid, fmt.Sprint("failed status callback with error ",
				err, " url ", dataCallRequest.StatusCallback))
		}
	}
}
