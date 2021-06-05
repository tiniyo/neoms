package managers

import (
	"fmt"
	"regexp"
	"github.com/neoms/config"
	"github.com/neoms/helper"
	"github.com/neoms/logger"
	"github.com/neoms/managers/callstats"
	"github.com/neoms/models"
)

func getXMLApplication(evHeaderMap map[string]string) *models.CallRequest {

	var data models.NumberAPIResponse
	phoneNumber := evHeaderMap["Variable_sip_req_user"]
	toPhoneNumber := evHeaderMap["Variable_sip_to_user"]
	callType := evHeaderMap["Variable_call_type"]
	callSid := evHeaderMap["Variable_call_sid"]
	callerId := evHeaderMap["Variable_sip_from_user"]
	fromUser := evHeaderMap["Variable_sip_from_user"]
	sipUser := evHeaderMap["Variable_sip_user"]
	parentCallSid := callSid
	url := ""
	key := phoneNumber

	if value := callstats.GetLocalCache(key); value != nil {
		data = value.(models.NumberAPIResponse)
	} else {

		if callType == "number" {
			logger.UuidInboundLog("Info", callSid, fmt.Sprint("sip call received on did number"))
			url = fmt.Sprintf("%s/%s", config.Config.Numbers.BaseUrl, numberSanity(phoneNumber))
		} else if callType == "number_tata" {
			logger.UuidInboundLog("Info", callSid, fmt.Sprint("sip call received on did number"))
			url = fmt.Sprintf("%s/%s", config.Config.Numbers.BaseUrl, numberSanity(toPhoneNumber))
			callType = "number"
			key = toPhoneNumber
		} else {
			logger.UuidInboundLog("Info", callSid, fmt.Sprint("sip call received from sip user"))
			url = fmt.Sprintf("%s/Endpoints/%s", config.Config.SipEndpoint.BaseUrl, sipUser)
			key = sipUser
			fromUser = sipUser
			callType = "sip"
		}

		if callType == "Wss" || callType == "wss" || callType == "Ws" {
			callType = "wss"
		} else if callType == "Sip" || callType == "sip" {
			callType = "sip"
		}

		logger.UuidInboundLog("Info", callSid, fmt.Sprint("get application url - ", url))

		statusCode, respBody, err := helper.Get(callSid, nil, url)
		if err != nil || statusCode != 200 {
			logger.UuidInboundLog("Err", callSid, fmt.Sprintf("url for response status code is  %d - %#v",
				statusCode, err))
			return nil
		}
		logger.UuidInboundLog("Info", callSid, fmt.Sprintf("get application response - %s", string(respBody)))
		err = json.Unmarshal(respBody, &data)
		if err != nil {
			logger.UuidInboundLog("Err", callSid, fmt.Sprintf("unmarshal application json failed, rejecting the calls %#v", err))
			return nil
		}

		callstats.SetLocalCache(key, data)
	}

	if data.AppResponse.AuthID == "" {
		logger.UuidInboundLog("Err", callSid, fmt.Sprintf("Phone-number %s is not attach "+
			"with any account, rejecting the calls  ", phoneNumber))
		return nil
	}

	if data.AppResponse.Application == (models.Application{}) {
		logger.UuidInboundLog("Err", callSid, fmt.Sprintf("Application is not attach with sip user or phone numebr"))
		return nil
	}

	if data.AppResponse.Application.InboundURL == "" {
		logger.UuidInboundLog("Err", callSid, fmt.Sprintf("Application is not attach with sip user or phone numebr"))
		return nil
	}

	pulse := float64(data.AppResponse.InitPulse)
	rate := pulse * data.AppResponse.Rate
	VendorAuthID := data.AppResponse.VendorAuthID
	if VendorAuthID == "" {
		VendorAuthID = "TINIYO1SECRET1AUTHID"
	}
	cr := models.CallRequest{}
	if data.AppResponse.Application.Name == "SIP_TRUNK" {
		logger.UuidInboundLog("Err", callSid, fmt.Sprintf("Application is SIP_TRUNK, We are not going to check for callerid"))
		cr.IsCallerId = "false"
		cr.SipTrunk = "true"
	}
	cr.CallSid = callSid
	cr.Sid = callSid
	cr.From = fromUser
	cr.ParentCallSid = parentCallSid
	cr.To = phoneNumber
	cr.CallResponse.Direction = "inbound"
	cr.Callback.Direction = "inbound"
	cr.Rate = rate
	cr.SrcDirection = "inbound"
	cr.SrcType = callType
	cr.CallerId = callerId
	cr.VendorAuthID = VendorAuthID
	cr.AccountSid = data.AppResponse.AuthID
	cr.Pulse = data.AppResponse.InitPulse
	cr.Url = data.AppResponse.Application.InboundURL
	cr.Method = data.AppResponse.Application.InboundMethod
	cr.StatusCallback = data.AppResponse.Application.StatusCallback
	cr.StatusCallbackMethod = data.AppResponse.Application.StatusCallbackMethod
	cr.StatusCallbackEvent = data.AppResponse.Application.StatusCallbackEvent
	cr.FallbackUrl = data.AppResponse.Application.FallbackUrl
	cr.FallbackMethod = data.AppResponse.Application.FallbackMethod
	cr.Host = data.AppResponse.Host
	if data.AppResponse.Host == "" {
		cr.Host = "tiniyo.com"
	}

	if callType == "number" {
		exportVars := "'tiniyo_accid,parent_call_sid,parent_call_uuid,tiniyo_did_number'"
		authIdSet := fmt.Sprintf("^^:tiniyo_accid=%s:tiniyo_rate=%f:"+
			"tiniyo_pulse=%d:"+
			"call_sid=%s:"+
			"call_type=Number:"+
			"parent_call_sid=%s:"+
			"parent_call_uuid=%s:"+
			"tiniyo_did_number=%s:"+
			"tiniyo_host=%s:"+
			"export_vars=%s", cr.AccountSid,
			cr.Rate, cr.Pulse, cr.Sid, cr.ParentCallSid,
			cr.ParentCallSid, numberSanity(phoneNumber), data.AppResponse.Host, exportVars)
		_ = MsAdapter.MultiSet(cr.CallSid, authIdSet)
	} else {
		exportVars := "'tiniyo_accid,parent_call_sid,parent_call_uuid'"
		authIdSet := fmt.Sprintf("^^:tiniyo_accid=%s:tiniyo_rate=%f:"+
			"tiniyo_pulse=%d:"+
			"call_sid=%s:"+
			"call_type=Sip:"+
			"parent_call_uuid=%s:"+
			"parent_call_sid=%s:"+
			"tiniyo_host=%s:"+
			"export_vars=%s", cr.AccountSid,
			cr.Rate, cr.Pulse, cr.Sid, cr.ParentCallSid, cr.ParentCallSid, data.AppResponse.Host, exportVars)
		_ = MsAdapter.MultiSet(cr.CallSid, authIdSet)
	}
	return &cr
}

func numberSanity(number string) string {
	reg, err := regexp.Compile("[^0-9]+")
	if err != nil {
		return number
	}
	return reg.ReplaceAllString(number, "")
}
