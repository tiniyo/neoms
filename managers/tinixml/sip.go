package tinixml

import (
	"encoding/json"
	"fmt"
	"github.com/beevik/etree"
	"github.com/neoms/config"
	"github.com/neoms/helper"
	"github.com/neoms/logger"
	"github.com/neoms/models"
	"strings"
)

/*
	<Response>
    	<Dial>
			<User>sip:chandra@phone.tiniyo.com</User>
    	</Dial>
	</Response>
*/
/*
	<Response>
    	<Dial>
			<Sip>sip:chandra@phone.tiniyo.com</Sip>
    	</Dial>
	</Response>
*/
func ProcessSip(dialVars, sipDestination string) string {
	toUser := strings.Split(sipDestination, "@")[0]
	sipTo := strings.Split(toUser, ":")
	if sipTo[0] == "sip" {
		toUser = sipTo[1]
	} else {
		toUser = sipTo[0]
	}

	dialNode := ""

	if locationData := locationLookup(sipDestination); locationData != "" {
		if strings.Contains(locationData, "transport=ws") {
		dialNode = fmt.Sprintf("[%s,sip_ignore_183nosdp=true," +
			"webrtc_enable_dtls=true,media_webrtc=true,sip_h_X-Tiniyo-Sip=%s," +
			"absolute_codec_string='opus@20i,PCMU@20i',ignore_early_media=true,"+
			"call_type=Sip,sip_h_X-Tiniyo-Phone=user]sofia/gateway/pstn_trunk/%s",
			dialVars, sipDestination,toUser)
		} else {
			dialNode = fmt.Sprintf("[%s,absolute_codec_string='PCMU,PCMA'," +
				"sip_ignore_183nosdp=true,ignore_early_media=true,"+
				"call_type=Sip,sip_h_X-Tiniyo-Phone=user,sip_h_X-Tiniyo-Sip=%s]sofia/gateway/pstn_trunk/%s",
				dialVars, sipDestination, toUser)
		}
	} else {
		dialNode = fmt.Sprintf("[%s,ignore_early_media=true,"+
			"sip_h_X-Tiniyo-Gateway=%s,absolute_codec_string='PCMU,PCMA',call_type=Sip,"+
			"sip_h_X-Tiniyo-Phone=sip,sip_h_X-Tiniyo-Sip=%s]sofia/gateway/pstn_trunk/%s",
			dialVars, sipDestination, sipDestination,toUser)
	}
	return dialNode
}

func ProcessSipAttr(data *models.CallRequest, child *etree.Element) {
	data.DialSipAttr.DialSipMethod = "POST"
	data.DialSipAttr.DialSipStatusCallbackEvent = "completed"
	data.DialSipAttr.DialSipStatusCallbackMethod = "POST"
	data.DestType = "sip"
	for _, attr := range child.Attr {
		switch attr.Key {
		case "method":
			data.DialSipAttr.DialSipMethod = attr.Value
		case "password":
			data.DialSipAttr.DialSipPassword = attr.Value
		case "url":
			data.DialSipAttr.DialSipUrl = attr.Value
		case "statusCallback":
			data.DialSipAttr.DialSipStatusCallback = attr.Value
		case "statusCallbackEvent":
			data.DialSipAttr.DialSipStatusCallbackEvent = attr.Value
		case "statusCallbackMethod":
			data.DialSipAttr.DialSipStatusCallbackMethod = attr.Value
		case "username":
			data.DialSipAttr.DialSipUsername = attr.Value
		default:
			logger.UuidLog("Err", data.ParentCallSid, fmt.Sprint("Attribute not supported - ", attr.Key))
		}
	}
}

/*
	Location lookup will be in kamgo instead of going to public
 */
func locationLookup(sipUser string) string {
	location := models.SipLocation{}
	url := fmt.Sprintf("%s/v1/Subscribers/%s/Locations", config.Config.Kamgo.BaseUrl, sipUser)
	statusCode, respBody, err := helper.Get(sipUser, nil, url)
	if err != nil || statusCode != 200 {
		return ""
	}
	if err = json.Unmarshal(respBody, &location);err != nil {
		return ""
	}
	return location.Contact
}
