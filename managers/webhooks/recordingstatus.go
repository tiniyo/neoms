package webhooks

import (
	"fmt"
	"github.com/neoms/helper"
	"github.com/neoms/logger"
	"github.com/neoms/models"
	"strings"
)

func ProcessRecordingStatusCallbackUrl(dataCallRequest models.CallRequest, state string) {
	if dataCallRequest.RecordingStatusCallback == ""{
		return
	}
	callSid := dataCallRequest.CallSid
	if callSid == ""{
		callSid = dataCallRequest.Sid
	}
	var err error
	dataMap := make(map[string]interface{})

	dataCallRequest.RecordCallback.RecordCallSid = dataCallRequest.ParentCallSid
	if dataCallRequest.RecordingStatusCallbackEvent == ""{
		dataCallRequest.RecordingStatusCallbackEvent = "completed"
	}

	if dataCallRequest.RecordCallback.RecordingDuration == "0" && state == "completed" {
		state = "absent"
		dataCallRequest.RecordCallback.RecordingStatus = "absent"
	}
	if callbackByte, err := json.Marshal(dataCallRequest.RecordCallback); err == nil {
		if err := json.Unmarshal(callbackByte, &dataMap); err != nil {
			logger.UuidLog("Err", callSid, fmt.Sprint("update issue - ", err))
		}
	}
	//get the child uuid and get the dial recordcallback
	if strings.Contains(dataCallRequest.RecordingStatusCallbackEvent, state) &&
		len(dataCallRequest.RecordingStatusCallback) > 0 && dataCallRequest.RecordingStatusCallbackMethod == "GET" {
		if _, _, err = helper.Get(callSid,dataMap, dataCallRequest.RecordingStatusCallback); err != nil {
			logger.UuidLog("Err", callSid, fmt.Sprint("failed recording status callback with error ",
				err, " url ", dataCallRequest.RecordingStatusCallback))
		}
	} else if strings.Contains(dataCallRequest.RecordingStatusCallbackEvent, state) &&
		len(dataCallRequest.RecordingStatusCallback) > 0 {
		if _, _, err = helper.Post(callSid,dataMap, dataCallRequest.RecordingStatusCallback); err != nil {
			logger.UuidLog("Err", callSid, fmt.Sprint("failed recording status callback with error ",
				err, " url ", dataCallRequest.RecordingStatusCallback))
		}
	}
}
