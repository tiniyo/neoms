package tinixml

import (
	"fmt"
	"github.com/beevik/etree"
	"github.com/tiniyo/neoms/logger"
	"github.com/tiniyo/neoms/models"
)

/*
	<Response>
    	<Dial>
			<Number>1234567890</Number>
    	</Dial>
	</Response>
*/
/*
	<Response>
    	<Dial>
			<Number callerId="12345678">1234567890</Number>
    	</Dial>
	</Response>
*/

func ProcessNumber(dialVars, destination string) string {
	dialNode := fmt.Sprintf("[%s,ignore_early_media=true,absolute_codec_string='PCMU,PCMA'," +
		"call_type=Number]sofia/gateway/pstn_trunk/%s", dialVars, destination)
	return dialNode
}

func ProcessNumberAttr(data *models.CallRequest, child *etree.Element)  {
	data.DialNumberAttr.DialNumberMethod = "POST"
	data.DialNumberAttr.DialNumberStatusCallbackEvent = "completed"
	data.DialNumberAttr.DialNumberStatusCallbackMethod = "POST"
	data.DestType = "number"
	for _, attr := range child.Attr {
		switch attr.Key {
		case "method":
			data.DialNumberAttr.DialNumberMethod = attr.Value
		case "sendDigits":
			data.DialNumberAttr.DialNumberSendDigits= attr.Value
		case "url":
			data.DialNumberAttr.DialNumberUrl = attr.Value
		case "statusCallback":
			data.DialNumberAttr.DialNumberStatusCallback = attr.Value
		case "statusCallbackEvent":
			data.DialNumberAttr.DialNumberStatusCallbackEvent = attr.Value
		case "statusCallbackMethod":
			data.DialNumberAttr.DialNumberStatusCallbackMethod = attr.Value
		case "byoc":
			data.DialNumberAttr.DialNumberByoc = attr.Value
		default:
			logger.UuidLog("Err", data.ParentCallSid, fmt.Sprint("Attribute not supported - ", attr.Key))
		}
	}
}
