package tinixml

import (
	"github.com/beevik/etree"
	"github.com/tiniyo/neoms/adapters"
	"github.com/tiniyo/neoms/logger"
)

/*
	<Response>
    	<Reject reason="busy"></Reject>
	</Response>
*/

func ProcessReject(msAdapter *adapters.MediaServer, uuid string, element *etree.Element) error {
	reason := "CALL_REJECTED"
	for _, attr := range element.Attr {
		logger.Logger.Debug("ATTR: %s=%s\n", attr.Key, attr.Value)
		if attr.Key == "reason" && attr.Value == "busy" {
			reason = "USER_BUSY"
		}
	}
	err := (*msAdapter).CallHangupWithReason(uuid, reason)
	return err
}
