package webhooks

import (
	"fmt"
	"github.com/neoms/helper"
	"github.com/neoms/logger"
	"github.com/neoms/models"
	"strings"
)

func ProcessDialNumberStatusCallbackUrl(dataCallRequest models.CallRequest, state string) {
	if dataCallRequest.DialNumberStatusCallback == "" {
		return
	}
	var err error
	callSid := dataCallRequest.CallSid
	dataMap := make(map[string]interface{})
	if callbackByte, err := json.Marshal(dataCallRequest.Callback); err == nil {
		if err := json.Unmarshal(callbackByte, &dataMap); err != nil {
			logger.UuidLog("Err", callSid, fmt.Sprint("call back map conversion issue - ", err))
		}
	}
	if strings.Contains(dataCallRequest.DialNumberStatusCallbackEvent, state) &&
		len(dataCallRequest.DialNumberStatusCallback) > 0 && dataCallRequest.DialNumberStatusCallbackMethod == "GET" {
		_, _, err = helper.Get(callSid,dataMap, dataCallRequest.DialNumberStatusCallback)
		if err != nil {
			logger.UuidLog("Err", callSid, fmt.Sprint("failed status callback with error ",
				err, " url ", dataCallRequest.DialNumberStatusCallback))
		}
	} else if strings.Contains(dataCallRequest.DialNumberStatusCallbackEvent, state) &&
		len(dataCallRequest.DialNumberStatusCallback) > 0 {
		_, _, err = helper.Post(callSid,dataMap, dataCallRequest.DialNumberStatusCallback)
		if err != nil {
			logger.UuidLog("Err", callSid, fmt.Sprint("failed status callback with error ",
				err, " url ", dataCallRequest.DialNumberStatusCallback))
		}
	}
}
