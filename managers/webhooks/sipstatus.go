package webhooks

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/tiniyo/neoms/helper"
	"github.com/tiniyo/neoms/logger"
	"github.com/tiniyo/neoms/models"
	"strings"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func ProcessDialSipStatusCallbackUrl(dataCallRequest models.CallRequest, state string) {
	if dataCallRequest.DialSipStatusCallback == "" {
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
	if strings.Contains(dataCallRequest.DialSipStatusCallbackEvent, state) &&
		dataCallRequest.DialSipStatusCallbackMethod == "GET" {
		_, _, err = helper.Get(callSid,dataMap, dataCallRequest.DialSipStatusCallback)
		if err != nil {
			logger.UuidLog("Err", callSid, fmt.Sprint("failed status callback with error ",
				err, " url ", dataCallRequest.DialSipStatusCallback))
		}

	} else if strings.Contains(dataCallRequest.DialSipStatusCallbackEvent, state) {
		_, _, err = helper.Post(callSid,dataMap, dataCallRequest.DialSipStatusCallback)
		if err != nil {
			logger.UuidLog("Err", callSid, fmt.Sprint("failed status callback with error ",
				err, " url ", dataCallRequest.DialSipStatusCallback))
		}

	}
}