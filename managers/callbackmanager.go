package managers

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/neoms/adapters"
	"github.com/neoms/adapters/callstate"
	"github.com/neoms/adapters/factory"
	"github.com/neoms/logger"
)

type CallBackManager struct {
	cs adapters.CallStateAdapter
}

type MediaServerCallBackHandler struct {
	Cs adapters.CallStateAdapter
}

var MsCB = new(MediaServerCallBackHandler)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var MsCallBackAdapter adapters.MediaServer

func (cm *CallBackManager) InitCallBackManager() {
	var err error
	cm.cs, err = callstate.NewCallStateAdapter()
	if err == nil {
		MsCB.Cs = cm.cs
	}
	MsCallBackAdapter = factory.GetMSInstance()
	_ = MsCallBackAdapter.InitializeCallbackMediaServers(MsCB)
}

func (msCB MediaServerCallBackHandler) CallBackMediaServerStatus(status int) error {
	if status > 0 {
		logger.Logger.Info(" MediaServer is Running ")
	} else {
		logger.Logger.Error(" MediaServer is Disconnected ")
	}
	return nil
}

func (msCB MediaServerCallBackHandler) CallBackOriginate(callSid string, evHeader []byte) error {
	logger.UuidLog("Info", callSid, "call initiated callback")
	go triggerCallBack("initiated", callSid, evHeader)
	return nil
}

func (msCB MediaServerCallBackHandler) CallBackProgressMedia(callSid string, evHeader []byte) error {
	logger.UuidLog("Info", callSid, "call ringing callback")
	go triggerCallBack("ringing", callSid, evHeader)
	return nil
}

func (msCB MediaServerCallBackHandler) CallBackHangup(uuid string) error {
	logger.Logger.WithField("uuid", uuid).Info("callback hangup received")
	return nil
}

func (msCB MediaServerCallBackHandler) CallBackHangupComplete(callSid string, evHeader []byte) error {
	logger.UuidLog("Info", callSid, "callback hangup complete")
	return triggerCallBack("completed", callSid, evHeader)
}

func (msCB MediaServerCallBackHandler) CallBackAnswered(callSid string, evHeader []byte) error {
	logger.UuidLog("Info", callSid, "call answer callback")
	go triggerCallBack("answered", callSid, evHeader)
	return nil
}

func (msCB MediaServerCallBackHandler) CallBackDTMFDetected(callSid string, evHeader []byte) error {
	dtmfMap := make(map[string]string)
	if err := json.Unmarshal(evHeader, &dtmfMap); err != nil {
		return err
	}
	callSid = dtmfMap["Unique-Id"]
	logger.UuidLog("Info", callSid, fmt.Sprint("dtmf detected - ", dtmfMap["Dtmf-Digit"]))
	go triggerDTMFCallBack(callSid, dtmfMap["Dtmf-Digit"])
	return nil
}

func (msCB MediaServerCallBackHandler) CallBackRecordingStart(callSid string, evHeader []byte) error {
	logger.Logger.WithField("uuid", callSid).Info("callback start recording received")
	go triggerRecordingCallBack("in-progress", callSid, evHeader)
	return nil
}

func (msCB MediaServerCallBackHandler) CallBackRecordingStop(callSid string, evHeader []byte) error {
	logger.Logger.WithField("uuid", callSid).Info("callback start recording received")
	go triggerRecordingCallBack("completed", callSid, evHeader)
	return nil
}

func (msCB MediaServerCallBackHandler) CallBackProgress(callSid string) error {
	logger.Logger.WithField("uuid", callSid).Info("callback progress received")
	return nil
}

/*
	Getting Park Events - Its useful to get the parked inbound call control from mediaserver to webfs
*/
func (msCB MediaServerCallBackHandler) CallBackPark(callSid string, evHeader []byte) error {
	logger.UuidLog("Info", callSid, "channel park callback")
	evHeaderMap := make(map[string]string)

	if err := json.Unmarshal(evHeader, &evHeaderMap); err != nil {
		logger.UuidInboundLog("Err", callSid, fmt.Sprint("error while unmarshal - ",
			err, " sending UNALLOCATED_NUMBER"))
		return nil
	}
	destType := evHeaderMap["Variable_tiniyo_destination"]

	switch destType {
	case "InboundXMLApp":
		logger.UuidInboundLog("Info", callSid, "inbound call received")
		callRequest := getXMLApplication(evHeaderMap)
		if callRequest == nil {
			logger.UuidInboundLog("Err", callSid, "not able to process the request, sending UNALLOCATED_NUMBER")
			return MsAdapter.CallHangupWithReason(callSid, "UNALLOCATED_NUMBER")
		}

		logger.UuidInboundLog("Info", callSid, fmt.Sprintf("callRequest: %#v\n", callRequest))

		/* Store Call State */
		jsonCallRequestByte, err := json.Marshal(callRequest)
		if err != nil {
			logger.UuidInboundLog("Err", callSid, "not able to process the request, sending UNALLOCATED_NUMBER")
			return MsAdapter.CallHangupWithReason(callSid, "UNALLOCATED_NUMBER")
		}
		if err = MsCB.Cs.Set(callRequest.Sid, jsonCallRequestByte); err != nil {
			logger.UuidInboundLog("Err", callSid, "not able to process the request, sending UNALLOCATED_NUMBER")
			return MsAdapter.CallHangupWithReason(callSid, "UNALLOCATED_NUMBER")
		}
		go triggerCallBack("initiated", callSid, evHeader)
		return nil
	case "Conf":
		authId := evHeaderMap["Variable_sip_h_x-tiniyo-authid"]
		confId := evHeaderMap["Variable_sip_h_x-tiniyo-conf"]
		confName := fmt.Sprintf("%s-%s@tiniyo+flags{moderator}", authId, confId)
		confBridgeCmd := fmt.Sprintf("%s", confName)
		_ = MsAdapter.ConfBridge(callSid, confBridgeCmd)
	default:

	}
	return nil
}

func (msCB MediaServerCallBackHandler) CallBackDestroy(uuid string) error {
	logger.Logger.Debug("CallDestroy Value :  uuid ", uuid)
	return nil
}

func (msCB MediaServerCallBackHandler) CallBackExecuteComplete(uuid string) error {
	logger.Logger.Debug("CallExecuteComplete Value :  uuid ", uuid)
	return nil
}

func (msCB MediaServerCallBackHandler) CallBackBridged(callSid string) error {
	logger.Logger.WithField("callSid", callSid).Debug(" CallBackBridged  - ")
	return nil
}

func (msCB MediaServerCallBackHandler) CallBackUnBridged(callSid string) error {
	logger.Logger.WithField("uuid", callSid).Debug(" CallBackUnBridged  - ")
	return nil
}

func (msCB MediaServerCallBackHandler) CallBackSessionHeartBeat(parentSid, callSid string) error {
	logger.Logger.WithField("uuid", callSid).
		WithField("parent", parentSid).Debug(" CallBackSessionHeartBeat  - ")
	parentSidRelationKey := fmt.Sprintf("parent:%s", parentSid)
	if callIds, err := MsCB.Cs.GetMembersScore(parentSidRelationKey); err != nil {
		logger.UuidLog("Err", callSid, fmt.Sprintf("Trouble while getting up child "+
			"and parent relationship in redis - %#v\n", err))
	} else {
		for callId, Score := range callIds {
			logger.Logger.WithField("uuid", callSid).
				WithField("parent", parentSid).Debug(" Sending heartbeat  - ")
			_, _ = MsCB.Cs.IncrKeyMemberScore(parentSidRelationKey, callId, 1)
			go makeHeartBeatRequest(callId, Score+1)
		}
	}
	return nil
}

func (msCB MediaServerCallBackHandler) CallBackMessage(callSid string) error {
	logger.Logger.WithField("uuid", callSid).Debug(" CallBackMessage  - ")
	return nil
}

func (msCB MediaServerCallBackHandler) CallBackCustom(callSid string) error {
	logger.Logger.WithField("uuid", callSid).Debug(" CallBackCustom  - ")
	return nil
}
