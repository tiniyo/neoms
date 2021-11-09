package tinixml

import (
	"errors"
	"fmt"
	"github.com/beevik/etree"
	"github.com/tiniyo/neoms/adapters"
	"github.com/tiniyo/neoms/logger"
	"github.com/tiniyo/neoms/managers/callstats"
	"github.com/tiniyo/neoms/models"
	"strconv"
)

func ProcessGather(msAdapter *adapters.MediaServer, data *models.CallRequest, child *etree.Element) (bool, error) {
	if data.Status != "in-progress" {
		_ = (*msAdapter).AnswerCall(data.CallSid)
	}
	callSid := data.CallSid
	ProcessGatherAttr(data, child)
	ProcessGatherChild(msAdapter, data, child)
	if timeout, err := strconv.Atoi(data.GatherTimeout); err == nil {
		ProcessPauseTime(timeout+5)
	}
	if currentState, err := callstats.GetCallBackDetailByUUID(callSid); currentState != nil && err ==nil {
		logger.UuidLog("Err", data.ParentCallSid, fmt.Sprint(currentState))
		if currentState.DtmfInputType == "dtmf" {
			return false, nil
		}else if data.GatherActionOnEmptyResult == "true"{
			//here we need to call dtmf timeout url
			//we will create a function in webhook and call it from here
			return false, errors.New("TIMEOUT")
		}else{
			return true, nil
		}
	}else {
		return false, nil
	}
}

func ProcessGatherAttr(data *models.CallRequest, child *etree.Element) {
	data.GatherAttr.GatherFinishOnKey = "#"
	data.GatherAttr.GatherTimeout = "5"
	data.GatherAttr.GatherMethod = "POST"
	data.GatherAttr.GatherActionOnEmptyResult = "false"
	for _, attr := range child.Attr {
		switch attr.Key {
		case "method":
			data.GatherAttr.GatherMethod = attr.Value
		case "action":
			data.GatherAttr.GatherAction = attr.Value
		case "finishOnKey":
			data.GatherAttr.GatherFinishOnKey = attr.Value
		case "numDigits":
			data.GatherAttr.GatherNumDigit = attr.Value
		case "actionOnEmptyResult":
			data.GatherAttr.GatherActionOnEmptyResult = attr.Value
		case "timeout":
			data.GatherAttr.GatherTimeout = attr.Value
		default:
			logger.UuidLog("Err", data.ParentCallSid, fmt.Sprint("Attribute not supported - ", attr.Key))
		}
	}
	/* Store Call State */
	err := callstats.SetCallDetailByUUID(data)
	if err != nil {
		logger.Logger.Error("SetCallState Failed", err)
	}
}

func ProcessGatherChild(msAdapter *adapters.MediaServer, data *models.CallRequest, child *etree.Element) {
	for _, dialChild := range child.ChildElements() {
		switch dialChild.Tag {
		case "Say", "Speak":
			_ = ProcessSpeak(msAdapter, *data, dialChild)
		case "Pause":
			ProcessPause(data.CallSid, dialChild)
		case "Play":
			_ = ProcessPlay(msAdapter, *data, dialChild)
		}
	}
}
