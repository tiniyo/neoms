package managers

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/tiniyo/neoms/adapters"
	"github.com/tiniyo/neoms/adapters/callstate"
	"github.com/tiniyo/neoms/logger"
)

type CallBackManager struct {
	callState    adapters.CallStateAdapter
	voiceAppMgr  VoiceAppManagerInterface
	heartBeatMgr HeartBeatManagerInterface
	webhookMgr   WebHookManagerInterface
}

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func (msCB *CallBackManager) InitCallBackManager() {
	msCB.callState, _ = callstate.NewCallStateAdapter()
	msCB.voiceAppMgr = NewVoiceAppManager()
	msCB.heartBeatMgr = NewHeartBeatManager(msCB.callState)
	xmlMgr := NewXmlManager(msCB.callState)
	msCB.webhookMgr = NewWebHookManager(msCB.callState, msCB.heartBeatMgr, xmlMgr)
}

func (msCB CallBackManager) CallBackMediaServerStatus(status int) error {
	if status > 0 {
		logger.Logger.Info(" MediaServer is Running ")
	} else {
		logger.Logger.Error(" MediaServer is Disconnected ")
	}
	return nil
}

func (msCB CallBackManager) CallBackOriginate(callSid string, evHeader []byte) error {
	logger.UuidLog("Info", callSid, "call initiated callback")
	err := msCB.webhookMgr.triggerCallBack("initiated", callSid, evHeader)
	if err != nil {
		logger.UuidLog("Err", callSid, fmt.Sprint("call initiated callback failed - ", err))
	}
	return err
}

func (msCB CallBackManager) CallBackProgressMedia(callSid string, evHeader []byte) error {
	logger.UuidLog("Info", callSid, "call ringing callback")
	err := msCB.webhookMgr.triggerCallBack("ringing", callSid, evHeader)
	if err != nil {
		logger.UuidLog("Err", callSid, fmt.Sprint("call ringing callback failed - ", err))
	}
	return err
}

func (msCB CallBackManager) CallBackHangup(uuid string) error {
	//logger.Logger.WithField("uuid", uuid).Info("callback hangup received")
	return nil
}

func (msCB CallBackManager) CallBackHangupComplete(callSid string, evHeader []byte) error {
	logger.UuidLog("Info", callSid, "callback hangup complete")
	err := msCB.webhookMgr.triggerCallBack("completed", callSid, evHeader)
	if err != nil {
		logger.UuidLog("Err", callSid, fmt.Sprint("call hangup complete callback failed - ", err))
	}
	return err
}

func (msCB CallBackManager) CallBackAnswered(callSid string, evHeader []byte) error {
	logger.UuidLog("Info", callSid, "call answer callback")
	err := msCB.webhookMgr.triggerCallBack("answered", callSid, evHeader)
	if err != nil {
		logger.UuidLog("Err", callSid, fmt.Sprint("call answer callback failed - ", err))
	}
	return err
}

func (msCB CallBackManager) CallBackDTMFDetected(callSid string, evHeader []byte) error {
	dtmfMap := make(map[string]string)
	if err := json.Unmarshal(evHeader, &dtmfMap); err != nil {
		return err
	}
	callSid = dtmfMap["Unique-Id"]
	logger.UuidLog("Info", callSid, fmt.Sprint("dtmf detected - ", dtmfMap["Dtmf-Digit"]))
	err := msCB.webhookMgr.triggerDTMFCallBack(callSid, dtmfMap["Dtmf-Digit"])
	if err != nil {
		logger.UuidLog("Err", callSid, fmt.Sprint("call dtmf callback failed - ", err))
	}
	return err
}

func (msCB CallBackManager) CallBackRecordingStart(callSid string, evHeader []byte) error {
	logger.Logger.WithField("uuid", callSid).Info("callback start recording received")
	err := msCB.webhookMgr.triggerRecordingCallBack("in-progress", callSid, evHeader)
	if err != nil {
		logger.UuidLog("Err", callSid, fmt.Sprint("call record start callback failed - ", err))
	}
	return err
}

func (msCB CallBackManager) CallBackRecordingStop(callSid string, evHeader []byte) error {
	logger.Logger.WithField("uuid", callSid).Info("callback stop recording received")
	go postRecordingData(callSid, evHeader)
	err := msCB.webhookMgr.triggerRecordingCallBack("completed", callSid, evHeader)
	if err != nil {
		logger.UuidLog("Err", callSid, fmt.Sprint("call record stop callback failed - ", err))
	}
	return err
}

func (msCB CallBackManager) CallBackProgress(callSid string) error {
	//	logger.Logger.WithField("uuid", callSid).Info("callback progress received")
	return nil
}

/*
	Getting Park Events - Its useful to get the parked inbound call control from mediaserver to webfs
*/
func (msCB CallBackManager) CallBackPark(callSid string, evHeader []byte) error {
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
		callRequest := msCB.voiceAppMgr.getXMLApplication(evHeaderMap)
		if callRequest == nil {
			logger.UuidInboundLog("Err", callSid, "not able to process the request, sending UNALLOCATED_NUMBER")
			return MsAdapter.CallHangupWithReason(callSid, "UNALLOCATED_NUMBER")
		}

		/* Store Call State */
		jsonCallRequestByte, err := json.Marshal(callRequest)
		if err != nil {
			logger.UuidInboundLog("Err", callSid, "not able to process the request, sending UNALLOCATED_NUMBER")
			return MsAdapter.CallHangupWithReason(callSid, "UNALLOCATED_NUMBER")
		}
		if err = msCB.callState.Set(callRequest.Sid, jsonCallRequestByte); err != nil {
			logger.UuidInboundLog("Err", callSid, "not able to process the request, sending UNALLOCATED_NUMBER")
			return MsAdapter.CallHangupWithReason(callSid, "UNALLOCATED_NUMBER")
		}

		err = msCB.webhookMgr.triggerCallBack("initiated", callSid, evHeader)
		if err != nil {
			logger.UuidInboundLog("Err", callSid, fmt.Sprint("error while sending initiated callback - ", err))
		}
		return err
	case "Conf":
		authId := evHeaderMap["Variable_sip_h_x-tiniyo-authid"]
		confId := evHeaderMap["Variable_sip_h_x-tiniyo-conf"]
		confName := fmt.Sprintf("%s-%s@tiniyo+flags{moderator}", authId, confId)
		confBridgeCmd := fmt.Sprintf("%s", confName)
		err := MsAdapter.ConfBridge(callSid, confBridgeCmd)
		if err != nil {
			return err
		}
	default:
		logger.UuidInboundLog("Info", callSid, "inbound call received")
	}
	return nil
}

func (msCB CallBackManager) CallBackDestroy(uuid string) error {
	//logger.Logger.Debug("CallDestroy Value :  uuid ", uuid)
	return nil
}

func (msCB CallBackManager) CallBackExecuteComplete(uuid string) error {
	//logger.Logger.Debug("CallExecuteComplete Value :  uuid ", uuid)
	return nil
}

func (msCB CallBackManager) CallBackBridged(callSid string) error {
	//logger.Logger.WithField("callSid", callSid).Debug(" CallBackBridged  - ")
	return nil
}

func (msCB CallBackManager) CallBackUnBridged(callSid string) error {
	//logger.Logger.WithField("uuid", callSid).Debug(" CallBackUnBridged  - ")
	return nil
}

func (msCB CallBackManager) CallBackSessionHeartBeat(parentSid, callSid string) error {
	logger.Logger.WithField("uuid", callSid).
		WithField("parent", parentSid).Debug(" CallBackSessionHeartBeat  - ")
	var callIds map[string]int64
	var err error
	parentSidRelationKey := fmt.Sprintf("parent:%s", parentSid)
	if callIds, err = msCB.callState.GetMembersScore(parentSidRelationKey); err != nil {
		logger.UuidLog("Err", callSid, fmt.Sprintf("Trouble while getting up child "+
			"and parent relationship in redis - %#v\n", err))
		return err
	}
	for callId, Score := range callIds {
		logger.UuidLog("Err", callSid, fmt.Sprintf("sending heartbeat"))
		if _, err = msCB.callState.IncrKeyMemberScore(parentSidRelationKey, callId, 1); err!=nil{
			//handle it later
		}
		err = msCB.heartBeatMgr.makeHeartBeatRequest(callId, Score+1)
		if err != nil {
			return err
		}
	}
	return nil
}

func (msCB CallBackManager) CallBackMessage(callSid string) error {
	//logger.Logger.WithField("uuid", callSid).Debug(" CallBackMessage  - ")
	return nil
}

func (msCB CallBackManager) CallBackCustom(callSid string) error {
	//logger.Logger.WithField("uuid", callSid).Debug(" CallBackCustom  - ")
	return nil
}
