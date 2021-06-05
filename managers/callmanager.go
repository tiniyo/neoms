package managers

import (
	"errors"
	"fmt"
	"github.com/neoms/adapters"
	"github.com/neoms/adapters/callstate"
	"github.com/neoms/adapters/factory"
	"github.com/neoms/helper"
	"github.com/neoms/logger"
	"github.com/neoms/managers/rateroute"
	"github.com/neoms/models"
	"strings"
)

const (
	GatewayDialString     = "sofia/gateway/pstn_trunk"
	SipGatwayDialString   = "sofia/gateway/pstn_trunk"
	DefaultVendorAuthId   = "TINIYO1SECRET1AUTHID"
)

type CallManager struct {
	callState adapters.CallStateAdapter
}

var MsAdapter adapters.MediaServer

func (cm *CallManager) InitCallManager() {
	cm.callState, _ = callstate.NewCallStateAdapter()
	MsAdapter = factory.GetMSInstance()
	var msCallbacker adapters.MediaServerCallbacker
	MsAdapter.InitializeCallbackMediaServers(msCallbacker)
}


/*
	415 - rateroute not found
 */

func (cm CallManager) CreateCall(cr *models.CallRequest) (*models.CallResponse, error) {
	var routingString = ""
	var destination = cr.To
	callResponse := models.CallResponse{}

	if cr.VendorAuthID == "" {
		cr.VendorAuthID = DefaultVendorAuthId
	}

	status, rateRoutes := rateroute.GetOutboundRateRoutes(cr.Sid, cr.VendorAuthID, cr.AccountSid, destination)
	if status == "failed" || rateRoutes == nil {
		logger.UuidLog("Err", cr.Sid, fmt.Sprintf("rateroute not found"))
		return nil, &models.RequestError{
			StatusCode: 415,
			Err:        errors.New("rates not set for destination"),
		}
	}


	if helper.IsSipCall(destination) {
		logger.Logger.Info("sip call processing, skipping route processing")
	} else {
		logger.Logger.Info("pstn call, termination route processing")
		if rateRoutes.Term == nil {
			logger.Logger.Error("No routes found, exit the call")
			return nil, &models.RequestError{
				StatusCode: 415,
				Err:        errors.New("routes not set for destination"),
			}
		}
		var routingTokenArray = helper.JwtTokenInfos{}
		for _, rt := range rateRoutes.Term {
			var routingToken = helper.JwtTokenInfo{}

			if rt.RemovePrefix != ""{
				cr.To = strings.TrimPrefix(cr.To, "+")
				cr.To = strings.TrimPrefix(cr.To,  rt.RemovePrefix)
			}
			if rt.FromRemovePrefix != ""{
				cr.From = strings.TrimPrefix(cr.From, "+")
				cr.From = strings.TrimPrefix(cr.From,  rt.FromRemovePrefix)
				cr.FromRemovePrefix = rt.FromRemovePrefix
			}
			if rt.TrunkPrefix != ""{
				cr.To = fmt.Sprintf("%s%s",rt.TrunkPrefix,cr.To)
			}

			if rt.SipPilotNumber != ""{
				cr.SipPilotNumber = rt.SipPilotNumber
			}
			if routingString == "" {
				routingString = fmt.Sprintf("sip:%s@%s", cr.To, rt.PrimaryIP)
			} else {
				routingString = fmt.Sprintf("%s^sip:%s@%s", routingString, cr.To, rt.PrimaryIP)
			}
			if rt.Username != "" {
				routingToken.Ip = rt.PrimaryIP
				routingToken.Username = rt.Username
				routingToken.Password = rt.Password
				routingTokenArray = append(routingTokenArray, routingToken)
			}
			if rt.FailoverIP != "" {
				routingString = fmt.Sprintf("%s^sip:%s@%s", routingString, cr.To, rt.FailoverIP)
				if rt.Username != "" {
					routingToken.Ip = rt.FailoverIP
					routingToken.Username = rt.Username
					routingToken.Password = rt.Password
					routingTokenArray = append(routingTokenArray, routingToken)
				}
			}
		}
	}

	logger.Logger.Info("Routing string for call is - ", routingString)
	pulse := rateRoutes.Orig.SubPulse
	rateInSecond := rateRoutes.Orig.Rate / 60
	rateInPulse := rateInSecond * float32(pulse)

	if cr.Record != "" {
		cr.RecordingSource = "OutboundAPI"
		if cr.RecordingTrack == "" {
			cr.RecordingTrack = "both"
		}
		if cr.RecordingChannels == "" {
			cr.RecordingChannels = "2"
		}
		if cr.RecordingStatusCallbackEvent == "" {
			cr.RecordingStatusCallbackEvent = "completed"
		}
	}

	cr.Rate = float64(rateInPulse)
	cr.Pulse = pulse
	cr.ParentCallSid = cr.Sid
	cr.Callback.CallSid = cr.Sid
	callResponse.Direction = "outbound-api"
	callResponse.PriceUnit = "USD"
	callResponse.To = cr.To
	callResponse.From = cr.From
	callResponse.Sid = cr.Sid
	cr.DestType = "api"
	callResponse.AccountSid = cr.AccountSid
	callResponse.APIVersion = "2010-04-01"
	callResponse.Status = "queued"
	callResponse.FromFormatted = cr.From
	callResponse.ToFormatted = cr.To
	callResponse.URI = fmt.Sprintf("/v1/Account/%s/Call/%s", cr.AccountSid, cr.Sid)
	cr.CallResponse = callResponse
	//get the json request of call request
	jsonCallRequestData, err := json.Marshal(cr)
	if err != nil {
		return nil, err
	}
	/* Store Call State */
	err = cm.callState.Set(cr.Sid, jsonCallRequestData)
	if err != nil {
		logger.Logger.Error("SetCallState Failed", err)
	}
	cr.SendDigits = helper.DtmfSanity(cr.SendDigits)
	originateStr := helper.GenDialString(cr)

	cmd := ""

	toUser := strings.Split(destination, "@")[0]
	sipTo := strings.Split(toUser, ":")
	if sipTo[0] == "sip" {
		toUser = sipTo[1]
	} else {
		toUser = sipTo[0]
	}

	if routingString != "" {
		cmd = fmt.Sprintf("bgapi originate {%s,sip_h_X-Tiniyo-Gateway=%s}%s/%s &park",
			originateStr, routingString, GatewayDialString, toUser)
	} else if helper.IsSipCall(destination) && strings.Contains(destination, "phone.tiniyo.com") {
		cmd = fmt.Sprintf("bgapi originate {%s,sip_h_X-Tiniyo-Gateway=%s,sip_h_X-Tiniyo-Phone=user}%s/%s &park",
			originateStr, destination, SipGatwayDialString, toUser)
	} else {
		cmd = fmt.Sprintf("bgapi originate {%s,sip_h_X-Tiniyo-Gateway=%s,sip_h_X-Tiniyo-Phone=sip}%s/%s &park",
			originateStr, destination, SipGatwayDialString, toUser)
	}

	logger.Logger.Debugln("Command : ", cmd)
	/* Make Call to the Call State */
	go func() {
		_ = MsAdapter.CallNewOutbound(cmd)
	}()
	return &callResponse, err
}

func (cm CallManager) UpdateCall(cr models.CallUpdateRequest) (*models.CallResponse, error) {
	callSid := cr.Sid
	logger.UuidLog("Info", callSid, "get current call status")
	data := models.CallRequest{}
	val, err := cm.callState.Get(callSid)

	if err == nil {
		if err := json.Unmarshal(val, &data); err != nil {
			logger.UuidLog("Err", callSid, fmt.Sprintf("no active call with callsid %s", err.Error()))
			return nil, err
		}

		if cr.StatusCallback != ""{
			data.StatusCallback = cr.StatusCallback
		}
		if cr.StatusCallbackMethod != ""{
			data.StatusCallback = cr.StatusCallbackMethod
		}
		if cr.Url != ""{
			data.Url = cr.Url
		}
		if cr.Method != ""{
			data.Method = cr.Method
		}
		if cr.FallbackUrl != ""{
			data.FallbackUrl = cr.FallbackUrl
		}
		if cr.FallbackMethod != ""{
			data.FallbackMethod = cr.FallbackMethod
		}

		if cr.Status == "canceled" && data.Status != "in-progress"{
		 	cm.DeleteCall(callSid)
			data.Status = "canceled"
		}else if cr.Status == "completed" {
			cm.DeleteCall(callSid)
			data.Status = "completed"
		}

		if dataByte, err := json.Marshal(data); err == nil {
			if err := MsCB.Cs.Set(callSid, dataByte); err != nil {
				logger.UuidLog("Err", callSid, fmt.Sprint(" triggerCallBack - callback state update issue - ", err))
			}
		}
		return &(data.CallResponse), nil
	}
	logger.UuidLog("Err", callSid, fmt.Sprintf("no active call with callsid %s", err.Error()))
	return nil, err
}

func (cm CallManager) GetCall(callSid string) (*models.CallResponse, error) {
	logger.UuidLog("Info", callSid, "get current call status")
	data := models.CallRequest{}
	val, err := cm.callState.Get(callSid)
	if err == nil {
		if err := json.Unmarshal(val, &data); err != nil {
			logger.UuidLog("Err", callSid, fmt.Sprintf("no active call with callsid %s", err.Error()))
			return nil, err
		}
		logger.UuidLog("Info", callSid, fmt.Sprint("current call status is - ", data))
		return &(data.CallResponse), nil
	}
	logger.UuidLog("Err", callSid, fmt.Sprintf("no active call with callsid %s", err.Error()))
	return nil, err
}

func (cm CallManager) DeleteCall(callSid string) {
	_ = MsAdapter.CallHangup(callSid)
}
func (cm CallManager) DeleteCallWithReason(callSid string, reason string) {
	_ = MsAdapter.CallHangupWithReason(callSid, reason)
}
