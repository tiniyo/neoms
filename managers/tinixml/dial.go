package tinixml

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/beevik/etree"
	uuid4 "github.com/satori/go.uuid"
	"github.com/tiniyo/neoms/adapters"
	"github.com/tiniyo/neoms/config"
	"github.com/tiniyo/neoms/helper"
	"github.com/tiniyo/neoms/logger"
	"github.com/tiniyo/neoms/managers/callstats"
	"github.com/tiniyo/neoms/managers/rateroute"
	"github.com/tiniyo/neoms/models"
)

//Default time-limit of calls is 240 minute or 4 hours

func ProcessDial(msAdapter *adapters.MediaServer, data models.CallRequest, child *etree.Element) (bool, bool, error) {
	var callReq models.CallRequest
	var err error

	parentCallSid := data.Sid
	callReq.SrcType = data.SrcType
	callReq.AccountSid = data.AccountSid
	callReq.From = data.From
	callReq.Host = data.Host
	callReq.IsCallerId = data.IsCallerId
	callReq.SrcDirection = data.SrcDirection
	callReq.CallerId = data.CallerId
	callReq.CallResponse.ParentCallSid = data.Sid
	callReq.ParentCallSid = data.Sid
	callReq.VendorAuthID = data.VendorAuthID
	callReq.FromRemovePrefix = data.FromRemovePrefix
	logger.UuidLog("Info", parentCallSid, fmt.Sprintf("IsCallerID is %s", callReq.IsCallerId))

	setDialAttribute(&callReq, parentCallSid, *child)

	setRingtone := fmt.Sprintf("^^!ringback=%s!instant_ringback=true", callReq.DialRingTone)
	_ = (*msAdapter).MultiSet(parentCallSid, setRingtone)

	/*
		AnswerOnBridge can be set on dial only, else behaviour is not defined
	*/
	data.Status = callstats.GetLiveCallStatus(parentCallSid)
	logger.UuidLog("Info", parentCallSid, fmt.Sprint("Before dial - call status is - ", data.Status))
	if data.Status != "in-progress" && callReq.DialAnswerOnBridge == "false" {
		logger.UuidLog("Info", parentCallSid, fmt.Sprint("answering the call"))
		if err := (*msAdapter).AnswerCall(data.CallSid); err != nil {
			return true, false, err
		}
	} else if data.Status != "in-progress" && callReq.DialAnswerOnBridge == "true" {
		logger.UuidLog("Info", parentCallSid, fmt.Sprint("pre answer call"))
		if err := (*msAdapter).PreAnswerCall(data.CallSid); err != nil {
			return true, false, err
		}
	}
	/*
		Rates not found for given phone number,exit from call
	*/
	dialString := ""
	confName := ""
	if dialString, confName, err = processDial(&callReq, parentCallSid, *child); err != nil {
		//process errors here
		if re, ok := err.(*models.RequestError); ok {
			if re.NestedDialElement() {
				logger.UuidLog("Info", parentCallSid, "dial element is nested with nouns- trying on nouns")
				dialString, confName, err = processDialChildes(&callReq, parentCallSid, *child)
			} else if re.RatingRoutingMissing() {
				logger.UuidLog("Info", parentCallSid, "rate-routes not found for dial, exiting")
				err = ProcessHangupWithTiniyoReason(msAdapter, data.CallSid, "Rates_Not_Found")
				return false, false, err
			} else if re.BadCallerID() {
				logger.UuidLog("Info", parentCallSid, "caller-id is not whitelisted, exiting")
				err = ProcessHangupWithTiniyoReason(msAdapter, data.CallSid, "CallerId_Not_Whitelisted")
				return false, false, err
			} else {
				logger.UuidLog("Info", parentCallSid, "errors are not defined, exiting")
				err = ProcessHangupWithTiniyoReason(msAdapter, data.CallSid, "Unknown_Reason")
				return false, false, err
			}
		}
	}

	if re, ok := err.(*models.RequestError); ok {
		if re.RatingRoutingMissing() {
			logger.UuidLog("Info", parentCallSid, "rate-routes not found for dial element, exiting")
			err = ProcessHangupWithTiniyoReason(msAdapter, data.CallSid, "Rates_Not_Found")
			return false, false, err
		} else if re.BadCallerID() {
			logger.UuidLog("Info", parentCallSid, "caller-id is not whitelisted for dial element, exiting")
			err = ProcessHangupWithTiniyoReason(msAdapter, data.CallSid, "CallerId_Not_Whitelisted")
			return false, false, err
		} else {
			logger.UuidLog("Info", parentCallSid, "errors are not defined for dial element, exiting")
			err = ProcessHangupWithTiniyoReason(msAdapter, data.CallSid, "Unknown_Reason")
			return false, false, err
		}
	}

	/*
		dial string and conference name not found, exit from call
	*/

	/* I moved this code as not matching up with twilio
	else if confName != "" && dialString != "" {
	logger.UuidLog("Info", parentCallSid, fmt.Sprintf("conference name is %s and dial string is %s",
		confName, dialString))
	confBridgeCmd := fmt.Sprintf("%s bgdial %s", confName, dialString)
	err = (*msAdapter).ConfSetAutoCall(data.Sid, dialString)
	if err != nil {
		logger.UuidLog("Err", parentCallSid, err.Error())
		return false, false, err
	}
	err = (*msAdapter).ConfBridge(data.Sid, confBridgeCmd)*/

	if dialString == "" && confName == "" {
		logger.UuidLog("Err", parentCallSid, "Empty dial string and conference name")
		err = ProcessHangupWithTiniyoReason(msAdapter, data.CallSid, "EMPTY_DIAL_STRING")
	} else if confName != "" {
		err = (*msAdapter).ConfBridge(data.Sid, confName)
	} else {
		if callReq.DialNumberUrl == "" {
			err = (*msAdapter).CallBridge(data.Sid, dialString)
		} else {
			dialString := fmt.Sprintf("bgapi originate %s &park", dialString)
			logger.UuidLog("Err", parentCallSid, fmt.Sprintf("bgapi originate command for outbound call is %s", dialString))
			if err = (*msAdapter).CallNewOutbound(dialString); err != nil {
				err = ProcessHangupWithTiniyoReason(msAdapter, data.CallSid, "Unknown_Reason")
			}
			if callReq.DialAction == "" {
				return true, true, err
			}
		}
	}
	if err != nil {
		logger.UuidLog("Err", parentCallSid, err.Error())
	}
	if callReq.DialAction == "" && err == nil {
		return true, false, nil
	}
	return false, false, err
}

func setDialAttribute(data *models.CallRequest, uuid string, child etree.Element) {
	logger.UuidLog("Info", uuid, fmt.Sprint("dial elements are", child.Attr))
	//recording attribute of file
	data.DialAnswerOnBridge = "false"
	data.RecordingTrack = "both"
	data.DialTimeLimit = "14400"
	data.DialTimeout = "30"
	data.DialRingTone = "%(2000,4000,440.0,480.0)"
	for _, attr := range child.Attr {
		switch strings.ToLower(attr.Key) {
		case "action":
			data.DialAction = attr.Value
		case "answeronbridge":
			data.DialAnswerOnBridge = attr.Value
		case "callerid":
			data.DialCallerId = attr.Value
			data.CallerId = attr.Value
		case "callreason":
			data.CallReason = attr.Value
		case "hanguponstar":
			data.DialHangupOnStar = attr.Value
		case "method":
			data.DialMethod = attr.Value
		case "record":
			data.Record = attr.Value
			data.RecordingSource = "DialVerb"
		case "recordingstatuscallback":
			data.RecordingStatusCallback = attr.Value
		case "recordingstatuscallbackmethod":
			data.RecordingStatusCallbackMethod = attr.Value
		case "recordingstatuscallbackevent":
			data.RecordingStatusCallbackEvent = attr.Value
		case "ringtone":
			if helper.IsValidRingTone(attr.Value) {
				data.DialRingTone = fmt.Sprintf("${%s-ring}", attr.Value)
			}
			if strings.HasPrefix(attr.Value, "http") {
				data.DialRingTone = fmt.Sprintf("%s", attr.Value)
			}
		case "recordingtrack":
			data.RecordingTrack = attr.Value
		case "timelimit":
			if helper.IsValidTimeLimit(attr.Value) {
				data.DialTimeLimit = attr.Value
			}
		case "timeout":
			if timeout, err := strconv.Atoi(attr.Value); err == nil {
				if timeout > 600 {
					data.DialTimeout = "600"
				} else if timeout < 5 {
					data.DialTimeout = "5"
				}
			}
			data.DialTimeout = attr.Value
		case "trim":
			data.Trim = attr.Value
		default:
			logger.UuidLog("Err", uuid, fmt.Sprint("Attribute not supported - ", attr.Key))
		}
	}
	logger.UuidLog("Info", uuid, fmt.Sprint("dial variables are - ", data))
}

/*
	This code will handle if dial dont have any child
<Dial record="yes">12234</Dial>
*/
func processDial(data *models.CallRequest, uuid string, child etree.Element) (string, string, error) {
	var err error
	dialDest := ""
	dialNode := ""
	confName := ""
	dialVars := ""
	//this block for dial text with no element
	dialText := child.Text()
	if dialText == "" || len(child.ChildElements()) != 0 {
		return "", "", &models.RequestError{
			StatusCode: 405,
			Err:        errors.New("dial text is nil or no dial child"),
		}
	}

	logger.UuidLog("Info", uuid, fmt.Sprintf("dial element text found - %s ", dialText))
	callSid := uuid4.NewV4().String() //generating uuid for each dial element
	data.Sid = callSid
	data.CallSid = callSid

	if strings.HasPrefix(dialText, "sip:") || strings.HasPrefix(dialText, "Sip:") {
		data.DestType = "sip"
		dialDest = fmt.Sprintf("%s", dialText)
		logger.UuidLog("Info", uuid, fmt.Sprint("dial child is sip - ", dialDest))

		dialVars, err = getDialAttributeDialString(data, dialDest)
		if err != nil {
			return "", "", err
		}

		rateRoute := rateroute.GetSetOutboundRateRoutes(data, dialText)
		if rateRoute == nil {
			logger.UuidLog("Err", uuid, fmt.Sprintf("rateroute not found"))
			return "", "", &models.RequestError{
				StatusCode: 503,
				Err:        errors.New("rates not set for destination"),
			}
		}

		if rateRoute.FromRemovePrefix != "" {
			callerId := data.DialCallerId
			callerId = strings.TrimPrefix(callerId, "+")
			callerId = strings.TrimPrefix(callerId, rateRoute.FromRemovePrefix)
			data.DialCallerId = callerId
		}

		dialVars = fmt.Sprintf("%s,origination_caller_id_number=%s,origination_caller_id_name=%s", dialVars, data.DialCallerId, data.DialCallerId)

		if rateRoute.SipPilotNumber != "" {
			dialVars = fmt.Sprintf("%s,sip_h_X-Tiniyo-Pilot-Number=%s", dialVars, rateRoute.SipPilotNumber)
		}

		if rateRoute.RemovePrefix != "" {
			dialDest = strings.TrimPrefix(dialDest, "+")
			dialDest = strings.TrimPrefix(dialDest, rateRoute.RemovePrefix)
		}

		if rateRoute.TrunkPrefix != "" {
			dialDest = fmt.Sprintf("%s%s", rateRoute.TrunkPrefix, dialDest)
		}

		dialVars = fmt.Sprintf("%s,call_sid=%s,origination_uuid=%s,"+
			"tiniyo_rate=%f,tiniyo_pulse=%d", dialVars, callSid, callSid, rateRoute.PulseRate, rateRoute.Pulse)
		if rateRoute.RoutingUserAuthToken != "" {
			dialVars = fmt.Sprintf("%s,sip_h_X-Tiniyo-Token=%s", dialVars, rateRoute.RoutingUserAuthToken)
		}
		dialNode = ProcessSip(dialVars, dialDest)
	} else {
		data.DestType = "number"
		dialDest = fmt.Sprintf("%s", dialText)
		dialVars, err = getDialAttributeDialString(data, dialDest)
		if err != nil {
			return "", "", err
		}
		rateRoute := rateroute.GetSetOutboundRateRoutes(data, dialDest)
		if rateRoute == nil {
			logger.UuidLog("Err", uuid, fmt.Sprintf("rateroute not found"))
			return "", "", &models.RequestError{
				StatusCode: 503,
				Err:        errors.New("rates not set for destination"),
			}
		}

		if rateRoute.FromRemovePrefix != "" {
			callerId := data.DialCallerId
			callerId = strings.TrimPrefix(callerId, "+")
			callerId = strings.TrimPrefix(callerId, rateRoute.FromRemovePrefix)
			data.DialCallerId = callerId
		}

		dialVars = fmt.Sprintf("%s,origination_caller_id_number=%s,origination_caller_id_name=%s", dialVars, data.DialCallerId, data.DialCallerId)

		if rateRoute.SipPilotNumber != "" {
			dialVars = fmt.Sprintf("%s,sip_h_X-Tiniyo-Pilot-Number=%s", dialVars, rateRoute.SipPilotNumber)
		}

		if rateRoute.RemovePrefix != "" {
			dialDest = strings.TrimPrefix(dialDest, "+")
			dialDest = strings.TrimPrefix(dialDest, rateRoute.RemovePrefix)
		}

		if rateRoute.TrunkPrefix != "" {
			dialDest = fmt.Sprintf("%s%s", rateRoute.TrunkPrefix, dialDest)
		}
		dialVars = fmt.Sprintf("%s,call_sid=%s,origination_uuid=%s,"+
			"tiniyo_rate=%f,tiniyo_pulse=%d,"+
			"sip_h_X-Tiniyo-Gateway=%s",
			dialVars, callSid, callSid, rateRoute.PulseRate, rateRoute.Pulse, rateRoute.RoutingGatewayString)
		if rateRoute.RoutingUserAuthToken != "" {
			dialVars = fmt.Sprintf("%s,sip_h_X-Tiniyo-Token=%s", dialVars, rateRoute.RoutingUserAuthToken)
		}
		dialNode = ProcessNumber(dialVars, dialText)
	}
	logger.UuidLog("Info", uuid, fmt.Sprint("dial  string - ", dialVars))
	return dialNode, confName, nil
}

func processDialChildes(data *models.CallRequest, uuid string, child etree.Element) (string, string, error) {
	if child.ChildElements() == nil {
		return "", "", nil
	}
	dialNode := ""
	confName := ""
	/*
		We need to create a copy of data
	*/

	for _, dialChild := range child.ChildElements() {
		uuidGen := uuid4.NewV4().String()
		data.Sid = uuidGen
		data.CallSid = uuidGen
		switch dialChild.Tag {
		case "Sip", "sip":
			ProcessSipAttr(data, dialChild)
		case "Number":
			ProcessNumberAttr(data, dialChild)
		case "Conference":
			confName = ProcessConference(data.CallSid, dialChild, data.AccountSid)
			continue
		default:
			logger.UuidLog("Err", uuid, fmt.Sprintf("Dial Tag is not supported %s", dialChild.Tag))
		}

		dialDest := fmt.Sprintf("%s", dialChild.Text())
		dialVars, err := getDialAttributeDialString(data, dialDest)
		if err != nil {
			return "", "", err
		}

		rateRoute := rateroute.GetSetOutboundRateRoutes(data, dialDest)
		if rateRoute == nil {
			logger.UuidLog("Err", uuid, fmt.Sprintf("rateroute not found"))
			return "", "", &models.RequestError{
				StatusCode: 503,
				Err:        errors.New("rates not set for destination"),
			}
		}

		if rateRoute.FromRemovePrefix != "" {
			callerId := data.DialCallerId
			callerId = strings.TrimPrefix(callerId, "+")
			callerId = strings.TrimPrefix(callerId, rateRoute.FromRemovePrefix)
			data.DialCallerId = callerId
		}

		dialVars = fmt.Sprintf("%s,origination_caller_id_number=%s,origination_caller_id_name=%s", dialVars, data.DialCallerId, data.DialCallerId)

		if rateRoute.SipPilotNumber != "" {
			dialVars = fmt.Sprintf("%s,sip_h_X-Tiniyo-Pilot-Number=%s", dialVars, rateRoute.SipPilotNumber)
		}

		if rateRoute.RemovePrefix != "" {
			dialDest = strings.TrimPrefix(dialDest, "+")
			dialDest = strings.TrimPrefix(dialDest, rateRoute.RemovePrefix)
		}

		if rateRoute.TrunkPrefix != "" {
			dialDest = fmt.Sprintf("%s%s", rateRoute.TrunkPrefix, dialDest)
		}

		if dialChild.Tag == "User" || dialChild.Tag == "Sip" {
			logger.UuidLog("Info", uuid, fmt.Sprint("sip endpoints - ", dialDest))
			if net.ParseIP(dialDest) != nil {
				dialDest = fmt.Sprintf("sip:%s@%s", data.To, dialDest)
				logger.UuidLog("Info", uuid, fmt.Sprint("dial child is ip - ", dialDest))
			} else if !strings.HasPrefix(dialDest, "sip:") {
				dialDest = fmt.Sprintf("sip:%s", dialDest)
				logger.UuidLog("Info", uuid, fmt.Sprint("dial child is sip - ", dialDest))
			}
			dialVars = fmt.Sprintf("%s,call_sid=%s,origination_uuid=%s,"+
				"tiniyo_rate=%f,tiniyo_pulse=%d", dialVars, uuidGen, uuidGen, rateRoute.PulseRate, rateRoute.Pulse)
			if rateRoute.RoutingUserAuthToken != "" {
				dialVars = fmt.Sprintf("%s,sip_h_X-Tiniyo-Token=%s", dialVars, rateRoute.RoutingUserAuthToken)
			}
			if dialNode != "" {
				tempNode := ProcessSip(dialVars, dialDest)
				dialNode = fmt.Sprintf("%s,%s", dialNode, tempNode)
			} else {
				dialNode = ProcessSip(dialVars, dialDest)
			}
		} else { //PSTN Phone Number
			dialVars = fmt.Sprintf("%s,call_sid=%s,origination_uuid=%s,"+
				"tiniyo_rate=%f,tiniyo_pulse=%d,"+
				"sip_h_X-Tiniyo-Gateway=%s",
				dialVars, uuidGen, uuidGen, rateRoute.PulseRate, rateRoute.Pulse, rateRoute.RoutingGatewayString)
			if rateRoute.RoutingUserAuthToken != "" {
				dialVars = fmt.Sprintf("%s,sip_h_X-Tiniyo-Token=%s", dialVars, rateRoute.RoutingUserAuthToken)
			}
			if dialNode != "" {
				tempNode := ProcessNumber(dialVars, dialDest)
				dialNode = fmt.Sprintf("%s,%s", dialNode, tempNode)
			} else {
				dialNode = ProcessNumber(dialVars, dialDest)
			}
		}
	}

	logger.UuidLog("Info", uuid, fmt.Sprint("dial  string - ", dialNode))
	return dialNode, confName, nil
}

func getDialAttributeDialString(data *models.CallRequest, dest string) (string, error) {
	//recording
	dialVars := ""
	if helper.IsValidRecordValue(data.Record) {
		recordVars := ProcessRecordDialAttribute(*data)
		if recordVars != "" {
			dialVars = fmt.Sprintf("%s", recordVars)
		}
	}
	strMaxDuration := fmt.Sprintf("+%s", data.DialTimeLimit)
	retainDuration := fmt.Sprintf("'sched_hangup %s %s alotted_timeout'", strMaxDuration, data.CallSid)
	if data.ParentCallSid != "" && dialVars != "" {
		dialVars = fmt.Sprintf("%s,tiniyo_accid=%s,parent_call_uuid=%s,parent_call_sid=%s,"+
			"direction=outbound-call,"+
			"originate_timeout=%s,leg_timeout=%s,bridge_answer_timeout=%s,"+
			"ringback='%s',transfer_ringback='%s',api_on_answer_2=%s",
			dialVars, data.AccountSid, data.ParentCallSid, data.ParentCallSid,
			data.DialTimeout, data.DialTimeout, data.DialTimeout, data.DialRingTone, data.DialRingTone, retainDuration)
	} else if data.ParentCallSid != "" {
		dialVars = fmt.Sprintf("tiniyo_accid=%s,parent_call_sid=%s,parent_call_uuid=%s,direction=outbound-call,"+
			"leg_timeout=%s,originate_timeout=%s,bridge_answer_timeout=%s,ringback='%s',transfer_ringback='%s',api_on_answer_2=%s", data.AccountSid,
			data.ParentCallSid, data.ParentCallSid,
			data.DialTimeout, data.DialTimeout, data.DialTimeout,
			data.DialRingTone, data.DialRingTone, retainDuration)
	}

	if data.DialHangupOnStar == "true" {
		dialVars = fmt.Sprintf("%s,bridge_terminate_key=*", dialVars)
	}

	if data.DialNumberSendDigits != "" {
		dialVars = fmt.Sprintf("%s,execute_on_answer_2='send_dtmf %s'", dialVars, data.DialNumberSendDigits)
	}

	logger.UuidLog("Info", data.ParentCallSid, fmt.Sprint("dial caller id is ",
		data.DialCallerId, " callerId is ", data.CallerId, " from is ", data.From))

	if data.FromRemovePrefix != "" {
		data.From = fmt.Sprintf("%s%s", data.FromRemovePrefix, data.From)
	}
	switch data.DestType {
	case "sip", "Sip", "SIP":
		logger.UuidLog("Info", data.CallSid, "No need to check for callerId, "+
			"destination is sip call")
		if data.DialCallerId == "" && data.CallerId != "" {
			data.DialCallerId = data.CallerId
		} else if data.DialCallerId == "" && data.From != "" {
			data.DialCallerId = data.From
		}
	case "number", "Number", "NUMBER":
		if data.IsCallerId == "false" {
			logger.UuidLog("Info", data.Sid, fmt.Sprint("not checking the callerid, processing the call"))
			data.DialCallerId = data.CallerId
		} else if data.SrcDirection == "inbound" && data.DialCallerId == "" {
			logger.UuidLog("Info", data.Sid, fmt.Sprint("not checking the callerid, processing the call"))
			data.DialCallerId = data.CallerId
		} else if isValidCallerId(data.AccountSid, data.ParentCallSid, dest) {
			if data.CallerId != "" {
				data.DialCallerId = data.CallerId
			} else if data.From != "" {
				data.DialCallerId = data.From
			} else {
				data.DialCallerId = dest
			}
		} else if isValidCallerId(data.AccountSid, data.ParentCallSid, data.CallerId) {
			data.DialCallerId = data.CallerId
		} else if isValidCallerId(data.AccountSid, data.ParentCallSid, data.From) {
			data.DialCallerId = data.From
		} else if !isValidCallerId(data.AccountSid, data.ParentCallSid, data.DialCallerId) {
			if data.SrcType != "number" {
				return "", &models.RequestError{
					StatusCode: 400,
					Err:        errors.New("invalid caller-id"),
				}
			}
			data.DialCallerId = data.From
		}
		//caller handle
		if dialVars != "" {
			dialVars = fmt.Sprintf("%s,instant_ringback=true", dialVars)
		} else {
			dialVars = fmt.Sprintf("instant_ringback=true")
		}
	}
	return dialVars, nil
}

//we are allowing callerId from purchased phone number and from verified phone number only
func isValidCallerId(authSid, callSid, callerId string) bool {
	if callerId == "" {
		logger.UuidLog("Err", callSid, fmt.Sprint("callerid is not set, not processing the call"))
		return false
	}
	numberServiceUrl := fmt.Sprintf("%s/%s/CallerIds/%s", config.Config.Numbers.BaseUrl, authSid, helper.RemovePlus(callerId))
	status, body, err := helper.Get(callSid, nil, numberServiceUrl)
	if err != nil || status != 200 || body == nil {
		logger.UuidLog("Err", callSid, fmt.Sprintf("callerid %s is not valid, exit from call", callerId))
		return false
	}
	logger.UuidLog("Info", callSid, fmt.Sprint("caller-id is valid, processing the call"))
	return true
}
