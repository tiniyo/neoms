package managers

import (
	"fmt"
	"regexp"
	"github.com/tiniyo/neoms/constant"

	"github.com/tiniyo/neoms/config"
	"github.com/tiniyo/neoms/helper"
	"github.com/tiniyo/neoms/logger"
	"github.com/tiniyo/neoms/models"
)

type VoiceAppManagerInterface interface {
	getXMLApplication(evHeaderMap map[string]string) *models.CallRequest
}

type VoiceAppManager struct {

}

func NewVoiceAppManager() VoiceAppManagerInterface {
	return VoiceAppManager{
	}
}

func (vAppMgr VoiceAppManager) getXMLApplication(evHeaderMap map[string]string) *models.CallRequest {

	var data models.NumberAPIResponse
	phoneNumber := evHeaderMap["Variable_sip_req_user"]
	toPhoneNumber := evHeaderMap["Variable_sip_to_user"]
	if phoneNumber == "" {
		phoneNumber = toPhoneNumber
	}
	callType := evHeaderMap["Variable_call_type"]
	callSid := evHeaderMap["Variable_call_sid"]
	callerId := evHeaderMap["Variable_sip_from_user"]
	fromUser := evHeaderMap["Variable_sip_from_user"]
	sipUser := evHeaderMap["Variable_sip_user"]
	parentCallSid := callSid
	url := ""

	if callType == "number" {
		logger.UuidInboundLog("Info", callSid, fmt.Sprint("sip call received on did number"))
		url = fmt.Sprintf("%s/%s", config.Config.Numbers.BaseUrl, numberSanity(phoneNumber))
	} else if callType == "number_tata" {
		logger.UuidInboundLog("Info", callSid, fmt.Sprint("sip call received on did number"))
		url = fmt.Sprintf("%s/%s", config.Config.Numbers.BaseUrl, numberSanity(toPhoneNumber))
		callType = "number"
	} else {
		logger.UuidInboundLog("Info", callSid, fmt.Sprint("sip call received from sip user :", callType, " ", sipUser))
		url = fmt.Sprintf("%s/Endpoints/%s", config.Config.SipEndpoint.BaseUrl, sipUser)
		fromUser = sipUser
		//callType = "sip"
	}

	if callType == "Wss" || callType == "wss" || callType == "Ws" {
		callType = "wss"
	} else {
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
	if VendorAuthID == "" || len(VendorAuthID) < 12 {
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
	cr.Caller = callerId
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
		exportVars := constant.GetConstant("NumberExportFsVars").(string)
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
		exportVars := constant.GetConstant("ExportFsVars").(string)
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
