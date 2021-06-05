package managers

import (
	"fmt"
	"github.com/neoms/config"
	"github.com/neoms/helper"
	"github.com/neoms/logger"
	"github.com/neoms/models"
)

func enableHeartbeat(data models.CallRequest) {
	callSid := data.CallSid

	logger.UuidLog("Info", callSid, fmt.Sprintf("enabling hearbeat first time,parent call sid is %s ", callSid))
	parentSidRelationKey := fmt.Sprintf("parent:%s",data.ParentCallSid)
	if err := MsCB.Cs.AddSetMember(parentSidRelationKey, callSid); err != nil {
		logger.UuidLog("Err", callSid, fmt.Sprintf("Trouble while setting up child "+
			"and parent relationship in redis - %#v\n", err))
	}

	logger.UuidLog("Info", callSid, fmt.Sprintf("sending heartbeat with rate %f %d ", data.Rate, data.Pulse))

	rate := data.Rate
	pulse := data.Pulse
	accId := data.AccountSid
	if err := MsAdapter.EnableSessionHeartBeat(callSid, "1"); err != nil {
		logger.UuidLog("Err", callSid, fmt.Sprintf("Trouble while enabling session heartbeat - %#v\n", err))
	}
	go sendHeartBeat(accId, callSid, rate, pulse)
}

func makeHeartBeatRequest(callSid string, heartbeatCount int64) error {
	msg := fmt.Sprintf("heartbeat count is %d",heartbeatCount)
	logger.UuidLog("Info", callSid, msg)
	val, err := MsCB.Cs.Get(callSid)
	if err == nil {
		/* Get answer url and its method */
		var data models.CallRequest
		if err := json.Unmarshal(val, &data); err != nil {
			logger.Logger.WithField("uuid", callSid).Error(" error while unmarshal, heartbeat processing failed  - ", err)
			return err
		}
		rate := data.Rate
		pulse := data.Pulse
		accId := data.AccountSid
		if heartbeatCount >= pulse{
			go sendHeartBeat(accId, callSid, rate, pulse)
			parentSidRelationKey := fmt.Sprintf("parent:%s",data.ParentCallSid)
			_, _ = MsCB.Cs.IncrKeyMemberScore(parentSidRelationKey, callSid, -int(pulse))
			//reset the score here
		}
	}
	return nil
}

func sendHeartBeat(accId string, callSid string, rate float64, duration int64) {
	dataMap := make(map[string]interface{})

	hearBeatUrl := fmt.Sprintf("%s/%s/Heartbeat/%s", config.Config.Heartbeat.BaseUrl, accId, callSid)
	hbRequest := models.HeartBeatPrivRequest{AccountID: accId,
		CallID:   callSid,
		Rate:     rate,
		Pulse:    duration,
		Duration: duration}

	logger.UuidLog("Info", callSid, fmt.Sprint("heartbeat request - ", hbRequest, " heartbeat url - ", hearBeatUrl))

	if byteData, err := json.Marshal(hbRequest); err == nil {
		if err := json.Unmarshal(byteData, &dataMap); err != nil {
			logger.UuidLog("Err", callSid, fmt.Sprint("send heartbeat failed - ", err))
			return
		}
	} else {
		logger.UuidLog("Err", callSid, fmt.Sprint("send heartbeat failed - ", err))
		return
	}

	logger.UuidLog("Info", callSid, fmt.Sprint("heartbeat request before post - ", dataMap,
		" heartbeat url - ", hearBeatUrl))

	statusCode, _, err := helper.Post(callSid,dataMap, hearBeatUrl)
	if err != nil {
		logger.UuidLog("Err", callSid, fmt.Sprint("send heartbeat failed - ", err))
	} else if statusCode != 200 {
		logger.UuidLog("Err", callSid, fmt.Sprint("send heartbeat failed - ", statusCode))

	} else {
		logger.UuidLog("Info", callSid, "heartbeat success response received")
	}
	return
}
