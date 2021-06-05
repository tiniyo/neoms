package callstats

import (
	"encoding/json"
	"fmt"
	"github.com/patrickmn/go-cache"
	"github.com/neoms/adapters"
	"github.com/neoms/adapters/callstate"
	"github.com/neoms/logger"
	"github.com/neoms/models"
	"time"
)

type CallStatManager struct {
}

var appCache *cache.Cache
var callStatObj adapters.CallStateAdapter

func (cs *CallStatManager) InitCallStatManager() {
	callStatObj, _ = callstate.NewCallStateAdapter()
	appCache = cache.New(5*time.Minute, 10*time.Minute)
}

func SetCallDetailByUUID(cr *models.CallRequest) error {
	if cr.CallSid == "" {
		cr.CallSid = cr.Sid
	}
	//get the json request of call request
	jsonCallRequestData, err := json.Marshal(cr)
	if err != nil {
		return err
	}
	/* Store Call State */
	err = callStatObj.Set(cr.CallSid, jsonCallRequestData)
	if err != nil {
		logger.Logger.Error("SetCallState Failed", err)
		return err
	}
	return nil
}

func GetCallDetailByUUID(uuid string) (*models.CallRequest, error) {
	var data models.CallRequest
	if val, err := callStatObj.Get(uuid); err == nil {
		logger.Logger.WithField("uuid", uuid).Info("call details are - ", string(val))
		if err := json.Unmarshal(val, &data); err != nil {
			logger.Logger.WithField("uuid", uuid).Error("error while unmarshal  - ", err)
			return nil, err
		}
		return &data, nil
	}
	return nil, nil
}

func GetCallBackDetailByUUID(callSid string) (*models.Callback, error) {
	var data models.Callback
	statusCallbackKey := fmt.Sprintf("statusCallback:%s", callSid)
	if val, err := callStatObj.Get(statusCallbackKey); err == nil {
		logger.Logger.WithField("uuid", callSid).Info("call details are - ", string(val))
		if err := json.Unmarshal(val, &data); err != nil {
			logger.Logger.WithField("uuid", callSid).Error("error while unmarshal  - ", err)
			return nil, err
		}
		return &data, nil
	}
	return nil, nil
}

func GetLiveCallStatus(callSid string) string {
	var statusCallback models.Callback
	statusCallbackKey := fmt.Sprintf("statusCallback:%s", callSid)
	logger.UuidLog("Info", callSid, fmt.Sprintf("Getting current status callback with key - %s", statusCallbackKey))
	if currentState, err := callStatObj.Get(statusCallbackKey); err == nil {
		if err := json.Unmarshal(currentState, &statusCallback); err != nil {
			logger.UuidLog("Err", callSid, fmt.Sprintf("triggerCallBack - error while unmarshal  - %s", err.Error()))
			return "no_status"
		}
		return statusCallback.CallStatus
	}
	return "no_status"
}
