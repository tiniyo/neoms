package managers

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/neoms/logger"
	"github.com/neoms/managers/webhooks"
	"github.com/neoms/models"
	"strconv"
	"strings"
	"time"
)

var FreeswitchJson = jsoniter.Config{
	EscapeHTML:             true,
	SortMapKeys:            true,
	ValidateJsonRawMessage: true,
	TagKey:                 "freeswitch",
}.Froze()

func triggerCallBack(state string, callSid string, evHeader []byte) error {
	var statusCallback models.Callback
	var dataCallRequest models.CallRequest
	var err error
	callCompleted := false

	statusCallbackKey := fmt.Sprintf("statusCallback:%s", callSid)
	logger.UuidLog("Info", callSid, fmt.Sprintf("triggerCallBack - getting current status callback with key - %s", statusCallbackKey))
	if currentState, err := MsCB.Cs.Get(statusCallbackKey); err == nil {
		if err := json.Unmarshal(currentState, &statusCallback); err != nil {
			logger.UuidLog("Err", callSid, fmt.Sprintf("triggerCallBack - error while unmarshal  - %s", err.Error()))
		}
	}

	if currentState, err := MsCB.Cs.Get(callSid); err == nil {
		if err := json.Unmarshal(currentState, &dataCallRequest); err != nil {
			logger.UuidLog("Err", callSid, fmt.Sprintf("triggerCallBack - error while unmarshal  - %s", err.Error()))
		}
	}

	if err := FreeswitchJson.Unmarshal(evHeader, &statusCallback); err != nil {
		logger.UuidLog("Err", callSid, fmt.Sprintf("triggerCallBack - unmarshal failed from freeswitch data - call might fail with error - %#v", err.Error()))
		return err
	}

	updateSequenceNumber(&statusCallback)

	evHeaderMap := make(map[string]string)

	if err := json.Unmarshal(evHeader, &evHeaderMap); err != nil {
		logger.Logger.WithField("uuid", callSid).Error("triggerCallBack - error while unmarshal  - ", err)
	} else {
		if dataCallRequest.SrcType == "sip" || dataCallRequest.SrcType == "wss" {
			fromUri := evHeaderMap["Variable_sip_from_uri"]
			statusCallback.From = fmt.Sprintf("sip:%s", fromUri)
			caller := evHeaderMap["Variable_Caller-ANI"]
			statusCallback.Caller = caller
		}
		if dataCallRequest.DestType == "sip" || dataCallRequest.DestType == "wss" {
			statusCallback.To = evHeaderMap["Variable_sip_h_x-tiniyo-sip"]
			statusCallback.Called = statusCallback.To
		}
	}
	if dataCallRequest.To != "" && statusCallback.To == ""{
		statusCallback.To = dataCallRequest.To
	}

	if dataCallRequest.From != "" && statusCallback.From == ""{
		statusCallback.From = dataCallRequest.From
	}

	if statusCallback.To == "" {
		statusCallback.To = statusCallback.Called
	}

	if statusCallback.From == "" {
		statusCallback.From = statusCallback.Caller
	}

	dataCallRequest.Status = state
	statusCallback.ApiVersion = "2010-04-01"
	statusCallback.CallbackSource = "call-progress-events"
	statusCallback.ParentCallSid = dataCallRequest.ParentCallSid

	switch state {
	case "initiated":
		statusCallback.CallStatus = "initiated"
		statusCallback.AccountSid = dataCallRequest.AccountSid
		statusCallback.InitiationTime = statusCallback.Timestamp
		if dataCallRequest.StatusCallbackEvent == "" {
			dataCallRequest.StatusCallbackEvent = "completed"
		}
	case "ringing":
		statusCallback.RingTime = statusCallback.Timestamp
		statusCallback.CallStatus = "ringing"
	case "in-progress", "answered":
		statusCallback.AnswerTime = statusCallback.Timestamp
		statusCallback.CallStatus = "in-progress"
	case "busy", "failed", "no-answer", "completed":
		if statusCallback.CallStatus == "ORIGINATOR_CANCEL" {
			statusCallback.CallStatus = "canceled"
		} else if statusCallback.CallStatus == "USER_BUSY" {
			statusCallback.CallStatus = "busy"
		} else if statusCallback.CallStatus == "NORMAL_CLEARING" {
			statusCallback.CallStatus = "completed"
		} else {
			statusCallback.CallStatus = "failed"
		}
		statusCallback.HangupTime = statusCallback.Timestamp
		callCompleted = true
	}

	dataCallRequest.Callback = statusCallback

	if dataByte, err := json.Marshal(statusCallback); err == nil {
		if err := MsCB.Cs.Set(statusCallbackKey, dataByte); err != nil {
			logger.UuidLog("Err", callSid, fmt.Sprint(" triggerCallBack - callback state update issue - ", err))
		}
	}

	go func() {
		if state == "answered" {
			enableHeartbeat(dataCallRequest)
		}
		webhooks.ProcessStatusCallbackUrl(dataCallRequest, state)
		webhooks.ProcessDialSipStatusCallbackUrl(dataCallRequest, state)
		webhooks.ProcessDialNumberStatusCallbackUrl(dataCallRequest, state)
	}()

	if callCompleted && dataCallRequest.DialAction != "" && dataCallRequest.ParentCallSid != "" {
		processDialActionUrl(dataCallRequest)
	} else if state == "initiated" && dataCallRequest.Callback.Direction == "inbound" {
		_ = handleXmlUrl(dataCallRequest)
	} else if state == "answered" && dataCallRequest.Callback.Direction == "outbound-api" {
		_ = processParentRequest(dataCallRequest, "")
	} else if (state == "answered") && (dataCallRequest.DialNumberUrl != "" ||
		dataCallRequest.DialSipUrl != "") {
		processDialNounUrl(dataCallRequest)
	}

	if callCompleted {
		parentCallSidKey := fmt.Sprintf("intercept:%s", dataCallRequest.ParentCallSid)
		if err := MsCB.Cs.Set(parentCallSidKey, []byte(callSid)); err != nil {
			logger.UuidLog("Err", callSid, fmt.Sprintf("error while setting the intercept key - %s", err.Error()))
		}
		freeCallResource(callSid, parentCallSidKey, statusCallbackKey, dataCallRequest.ParentCallSid)
	}
	return err
}

func triggerRecordingCallBack(state string, callSid string, evHeader []byte) error {
	redirect := false
	logger.UuidLog("Err", callSid, fmt.Sprint("recording event to webhook"))
	recordJob := models.RecordJob{}
	var recordCallback models.RecordCallback
	var err error
	var dataCallRequest models.CallRequest

	recordingKey := fmt.Sprintf("recording:%s", callSid)

	if currentRecordingState, err := MsCB.Cs.Get(recordingKey); err == nil {
		if err := json.Unmarshal(currentRecordingState, &recordCallback); err != nil {
			logger.UuidLog("Err", callSid, fmt.Sprintf("error while unmarshal  - %s", err.Error()))
		}
	}

	if currentState, err := MsCB.Cs.Get(callSid); err == nil {
		logger.UuidLog("Info", callSid, fmt.Sprintf("current states from redis is - %s", currentState))
		if err := json.Unmarshal(currentState, &dataCallRequest); err != nil {
			logger.UuidLog("Err", callSid, fmt.Sprintf("error while unmarshal  - %s", err.Error()))
		}
	}

	parentCallSid := dataCallRequest.ParentCallSid

	if err := FreeswitchJson.Unmarshal(evHeader, &recordCallback); err != nil {
		logger.UuidLog("Err", callSid, err.Error())
		return err
	}
	recordCallback.RecordingStartTime = dataCallRequest.RecordingStartTime
	recordCallback.RecordingSid = callSid
	recordCallback.RecordCallSid = callSid
	recordCallback.RecordAccountSid = dataCallRequest.AccountSid
	switch state {
	case "in-progress":
		recordCallback.RecordingStartTime = recordCallback.RecordEventTimestamp
		recordCallback.RecordingStatus = "in-progress"
		if dataCallRequest.RecordingStatusCallbackEvent == "" {
			dataCallRequest.RecordingStatusCallbackEvent = "completed"
		}
		recordFile := fmt.Sprintf("tiniyo_recording_file=https://api.%s/v1/Accounts/%s/Recordings/%s.mp3",
			dataCallRequest.Host, recordCallback.RecordAccountSid, recordCallback.RecordingSid)
		if err := MsAdapter.Set(callSid, recordFile); err != nil {
			logger.UuidLog("Err", callSid, fmt.Sprintf("Not able to set the recording file name %s", recordFile))
		}
		if callSid != parentCallSid {
			if err := MsAdapter.Set(parentCallSid, recordFile); err != nil {
				logger.UuidLog("Err", callSid, fmt.Sprintf("Not able to set the recording file name on parent callsid %s", recordFile))
			}
		}

	case "completed":
		logger.UuidLog("Info", callSid, fmt.Sprintf("recording complete event is received %s and event %s",
			dataCallRequest.RecordingStatusCallback, dataCallRequest.RecordingStatusCallbackEvent))
		recordCallback.RecordingStatus = "completed"
		if dataCallRequest.RecordingSource == "RecordVerb" {
			if dataCallRequest.RecordAction == "" {
				//nothing to do here
			} else {
				dataCallRequest.Url = dataCallRequest.RecordAction
				dataCallRequest.Method = dataCallRequest.RecordMethod
			}
			redirect = true
		}
		recordJob.Name = "s3_upload"
		recordJob.T = time.Now().UnixNano()
		recordJob.ID = recordCallback.RecordCallSid
		recordJob.Args.JobID = int64(time.Now().Second())
		recordJob.Args.FilePath = recordCallback.RecordingUrl
		recordJob.Args.FileName = fmt.Sprint(recordCallback.RecordAccountSid, "-", recordCallback.RecordCallSid, ".mp3")
		if recordJobByte, err := json.Marshal(recordJob); err == nil {
			_ = MsCB.Cs.SetRecordingJob(recordJobByte)
		}
		recordFile := fmt.Sprintf("https://api.%s/v1/Accounts/%s/Recordings/%s.mp3", dataCallRequest.Host,
			recordCallback.RecordAccountSid, recordCallback.RecordingSid)
		recordCallback.RecordingUrl = recordFile
	}
	dataCallRequest.RecordCallback = recordCallback
	logger.UuidLog("Info", callSid, fmt.Sprint("sending updates to redis - recording callback ", recordCallback))

	webhooks.ProcessRecordingStatusCallbackUrl(dataCallRequest, state)

	if dataByte, err := json.Marshal(recordCallback); err == nil {
		_ = MsCB.Cs.Set(recordingKey, dataByte)
	}
	if redirect {
		_ = handleXmlUrl(dataCallRequest)
	}
	return err
}

/*
	In dtmf callback we are calling the statuscallback object only
	Make sure to modify only dtmf and staus callback only
*/
func triggerDTMFCallBack(callSid string, digit string) error {
	var dataCallRequest models.CallRequest
	var statusCallback models.Callback
	dtmfDone := false
	statusCallbackKey := fmt.Sprintf("statusCallback:%s", callSid)
	logger.UuidLog("Info", callSid, fmt.Sprintf("triggerDTMFCallBack - getting current status callback with key - %s", statusCallbackKey))
	if currentState, err := MsCB.Cs.Get(statusCallbackKey); err == nil {
		if err := json.Unmarshal(currentState, &statusCallback); err != nil {
			logger.UuidLog("Err", callSid, fmt.Sprintf("triggerDTMFCallBack - error while unmarshal  - %s", err.Error()))
		}
	}

	if currentState, err := MsCB.Cs.Get(callSid); err == nil {
		logger.UuidLog("Info", callSid, fmt.Sprintf("triggerDTMFCallBack - current states from redis is - %s", currentState))
		if err := json.Unmarshal(currentState, &dataCallRequest); err != nil {
			logger.UuidLog("Err", callSid, fmt.Sprintf("error while unmarshal  - %s", err.Error()))
		}
	}
	statusCallback.DtmfInputType = "dtmf"

	if dataCallRequest.GatherNumDigit != "" {
		if statusCallback.Digits == "" {
			statusCallback.Digits = fmt.Sprintf("%s", digit)
		} else {
			statusCallback.Digits = fmt.Sprintf("%s%s", statusCallback.Digits, digit)
		}
		numDigitReceived := len(statusCallback.Digits)
		if numDigitLimit, err := strconv.Atoi(dataCallRequest.GatherNumDigit); err == nil && numDigitLimit > 0 {
			if numDigitReceived >= numDigitLimit {
				dtmfDone = true
			}
		}
	} else if dataCallRequest.GatherFinishOnKey == digit {
		dtmfDone = true
	} else {
		if statusCallback.Digits == "" {
			statusCallback.Digits = fmt.Sprintf("%s", digit)
		} else {
			statusCallback.Digits = fmt.Sprintf("%s%s", statusCallback.Digits, digit)
		}
	}

	dataCallRequest.Callback = statusCallback

	//now digit pressed finished, start processing
	if dtmfDone {
		if dataCallRequest.GatherAction != "" {
			dataCallRequest.Url = dataCallRequest.GatherAction
			dataCallRequest.Method = dataCallRequest.GatherMethod
		}
		go func() {
			_ = handleXmlUrl(dataCallRequest)
		}()
		statusCallback.Digits = ""
	}

	if dataByte, err := json.Marshal(statusCallback); err == nil {
		if err := MsCB.Cs.Set(statusCallbackKey, dataByte); err != nil {
			logger.UuidLog("Err", callSid, fmt.Sprint(" triggerDTMFCallBack - callback state update issue - ", err))
		}
	}
	return nil
}

/*
	In dtmf callback we are calling the statuscallback object only
	Make sure to modify only dtmf and staus callback only
*/
func triggerDTMFTimeoutCallBack(callSid string) error {
	var dataCallRequest models.CallRequest
	var statusCallback models.Callback
	statusCallbackKey := fmt.Sprintf("statusCallback:%s", callSid)
	logger.UuidLog("Info", callSid, fmt.Sprintf("triggerDTMFCallBack - getting current status callback with key - %s", statusCallbackKey))
	if currentState, err := MsCB.Cs.Get(statusCallbackKey); err == nil {
		if err := json.Unmarshal(currentState, &statusCallback); err != nil {
			logger.UuidLog("Err", callSid, fmt.Sprintf("triggerDTMFCallBack - error while unmarshal  - %s", err.Error()))
		}
	}

	if currentState, err := MsCB.Cs.Get(callSid); err == nil {
		logger.UuidLog("Info", callSid, fmt.Sprintf("triggerDTMFCallBack - current states from redis is - %s", currentState))
		if err := json.Unmarshal(currentState, &dataCallRequest); err != nil {
			logger.UuidLog("Err", callSid, fmt.Sprintf("error while unmarshal  - %s", err.Error()))
		}
	}
	statusCallback.DtmfInputType = "dtmf"
	statusCallback.Digits = ""
	dataCallRequest.Callback = statusCallback

	if dataCallRequest.GatherAction != "" {
		dataCallRequest.Url = dataCallRequest.GatherAction
		dataCallRequest.Method = dataCallRequest.GatherMethod
	}
	go func() {
		_ = handleXmlUrl(dataCallRequest)
	}()
	statusCallback.Digits = ""
	return nil
}

func processParentRequest(data models.CallRequest, childState string) error {
	callSid := data.CallSid
	if data.SendDigits != "" && childState == "" {
		digits := data.SendDigits
		wSCount := strings.Count(digits, "w")
		wBCount := strings.Count(digits, "W")
		digitCount := len(data.SendDigits) - (wSCount + wBCount)
		sleepDuration := (float32(wSCount)*0.5 + float32(wBCount)*1 + float32(digitCount)*0.6) + 1
		time.Sleep(time.Duration(sleepDuration) * time.Second)
	}
	logger.UuidLog("Info", callSid, fmt.Sprint("now calling the "+
		"action url with direction - ", data.CallResponse.Direction))
	if (data.Callback.Direction == "outbound-api") ||
		(data.Callback.Direction == "outbound-call") && childState == "answered" ||
		(data.Callback.Direction == "inbound" && childState == "completed") {
		ProcessXmlResponse(data)
	}
	return nil
}

/*
	once dial call ends it will check for action url, if action url found it will get the xml
*/
func processDialActionUrl(dataCallRequest models.CallRequest) {
	callSid := dataCallRequest.CallSid
	if callState, err := MsCB.Cs.Get(dataCallRequest.ParentCallSid); err == nil {
		var parentData models.CallRequest
		if err := json.Unmarshal(callState, &parentData); err != nil {
			logger.UuidLog("Err", callSid, "error while unmarshal")
		} else {
			parentData.Url = dataCallRequest.DialAction
			parentData.Method = dataCallRequest.DialMethod
			parentData.DialCallStatus = dataCallRequest.CallStatus
			parentData.DialCallDuration = dataCallRequest.CallDuration
			parentData.DialCallSid = dataCallRequest.CallSid
			parentData.Callback.RecordingUrl = dataCallRequest.RecordCallback.RecordingUrl
			_ = processParentRequest(parentData, "completed")
		}
	}
}

func processDialNounUrl(dataCallRequest models.CallRequest) {
	/*
		Here we need to handle outbound call those have url set,we need to execute that
		url on outbound leg and then do the bridge call with parent call
	*/
	if dataCallRequest.DialSipUrl != "" {
		dataCallRequest.Url = dataCallRequest.DialSipUrl
		dataCallRequest.Method = dataCallRequest.DialSipMethod
	} else {
		dataCallRequest.Url = dataCallRequest.DialNumberUrl
		dataCallRequest.Method = dataCallRequest.DialNumberMethod
	}

	_ = processParentRequest(dataCallRequest, "answered")
}

func updateSequenceNumber(statusCallback *models.Callback) {
	callSid := statusCallback.CallSid
	if statusCallback.SequenceNumber == "" {
		statusCallback.SequenceNumber = "0"
	} else {
		seqNum, err := strconv.Atoi(statusCallback.SequenceNumber)
		if err != nil {
			seqNum = 0
		} else {
			seqNum = seqNum + 1
			statusCallback.SequenceNumber = strconv.Itoa(seqNum)
		}
	}
	logger.UuidLog("Info", callSid, fmt.Sprintf("sequence number updated"))
}

func freeCallResource(callSid, parentCallSidKey, statusCallbackKey, parentCallSid string) {
	time.Sleep(2 * time.Second)
	parentSidRelationKey := fmt.Sprintf("parent:%s",parentCallSid)
	recordingKey := fmt.Sprintf("recording:%s", callSid)
	_ = MsCB.Cs.Del(callSid)
	_ = MsCB.Cs.Del(parentCallSidKey)
	_ = MsCB.Cs.Del(statusCallbackKey)
	_ = MsCB.Cs.Del(recordingKey)
	_ = MsCB.Cs.DelKeyMember(parentSidRelationKey, callSid)
}
