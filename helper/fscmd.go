package helper

import (
	"fmt"
	"github.com/neoms/logger"
	"github.com/neoms/models"
	"strconv"
	"strings"
)

const (
	SipCallIdentityHeader = "sip_h_X-Tiniyo-Phone"
	ExportVars            = "'tiniyo_accid\\,tiniyo_rate\\,tiniyo_pulse\\,parent_call_uuid\\,parent_call_sid'"
	DefaultCodecs         = "'PCMU,PCMA'"
	DefaultCallerIdType   = "pid"
	InstantRingback       = "true"
	RingBack              = "'%(2000\\,4000\\,440\\,480)'"
)

func IsSipCall(dest string) bool {
	if strings.HasPrefix(dest, "sip:") {
		return true
	}
	return false
}


/*
	generating the dial string from vars
*/
func ConvertMapToDialString(dialVars map[string]string) (string, bool) {
	logger.Logger.Debug("In getDialString - ", dialVars)
	dialVarsCount := 0
	originateStr := ""
	isSipDestination := false
	for key, element := range dialVars {
		if key == SipCallIdentityHeader && element == "true" {
			isSipDestination = true
		}
		if dialVarsCount == 0 {
			originateStr = fmt.Sprintf("%s=%s", key, element)
		} else {
			originateStr = fmt.Sprintf("%s,%s=%s", originateStr, key, element)
		}
		dialVarsCount++
	}
	return originateStr, isSipDestination
}

/*
	creating map of dial vars
*/
func GenDialString(cr *models.CallRequest) string {
	originateVars := make(map[string]string)
	if strings.HasPrefix(cr.To, "sip:") || strings.HasPrefix(cr.To, "sips:") {
		originateVars[SipCallIdentityHeader] = "true"
	}
	if cr.MaxDuration > 0 {
		strMaxDuration := fmt.Sprintf("+%d", cr.MaxDuration)
		retainDuration := fmt.Sprintf("'sched_hangup %s %s alotted_timeout'", strMaxDuration, cr.Sid)
		originateVars["api_on_answer"] = retainDuration
	}

	if cr.SipPilotNumber != ""{
		originateVars["sip_h_X-Tiniyo-Pilot-Number"] = cr.SipPilotNumber
	}
	originateVars["originate_timeout"] = "65"
	originateVars["ignore_early_media"] = "ring_ready"
	if cr.Timeout != "" {
		if timeout, err := strconv.Atoi(cr.Timeout); err == nil && timeout < 600 {
			timeout = timeout + 5
			originateVars["originate_timeout"] = fmt.Sprintf("%d", timeout)
		}
	}
	originateVars["bridge_answer_timeout"] = originateVars["originate_timeout"]

	if cr.Record == "true" {
		recordDir := "/call_recordings"
		recordFile := fmt.Sprintf("%s/%s-%s.mp3", recordDir, cr.AccountSid, cr.Sid)
		recordString := fmt.Sprintf("'record_session %s'", recordFile)
		originateVars["media_bug_answer_req"] = "true"
		originateVars["execute_on_answer_1"] = recordString
		switch cr.RecordingTrack {
		case "inbound":
			originateVars["RECORD_READ_ONLY"] = "true"
		case "outbound":
			originateVars["RECORD_WRITE_ONLY"] = "true"
		default:
			if cr.RecordingChannels == "dual" {
				originateVars["RECORD_STEREO"] = "true"
			}
		}
	}

	if strings.HasPrefix(cr.To, "sip:") {
		originateVars["call_type"] = "Sip"
	} else {
		originateVars["call_type"] = "Number"
	}

	originateVars["ringback"] = RingBack
	originateVars["instant_ringback"] = InstantRingback
	strCallerID := fmt.Sprintf("%s", cr.From)
	originateVars["origination_caller_id_number"] = strCallerID
	originateVars["origination_caller_id_name"] = strCallerID
	originateVars["sip_cid_type"] = DefaultCallerIdType
	originateVars["absolute_codec_string"] = DefaultCodecs
	originateVars["call_sid"] = cr.Sid
	originateVars["origination_uuid"] = cr.Sid
	originateVars["parent_call_sid"] = cr.ParentCallSid
	originateVars["parent_call_uuid"] = cr.ParentCallSid
	originateVars["tiniyo_accid"] = cr.AccountSid
	originateVars["direction"] = "outbound-api"
	strRate := fmt.Sprintf("%f", cr.Rate)
	originateVars["tiniyo_rate"] = strRate
	if cr.SendDigits != "" {
		originateVars["execute_on_answer_2"] = fmt.Sprintf("'send_dtmf %s'", cr.SendDigits)
	}
	strPulse := fmt.Sprintf("%d", cr.Pulse)
	originateVars["tiniyo_pulse"] = strPulse
	originateVars["export_vars"] = ExportVars
	originateString, _ := ConvertMapToDialString(originateVars)
	return originateString
}