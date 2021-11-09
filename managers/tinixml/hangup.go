package tinixml

import (
	"github.com/beevik/etree"
	"github.com/tiniyo/neoms/adapters"
	"github.com/tiniyo/neoms/logger"
)

/*
   <Response>
   <Hangup reason="busy"></Hangup>
   </Response>
*/

func ProcessHangup(msAdapter *adapters.MediaServer, uuid string, element *etree.Element) error {
	reason := "CALL_REJECTED"
	for _, attr := range element.Attr {
		logger.Logger.Debug("ATTR: %s=%s\n", attr.Key, attr.Value)
		if attr.Key == "reason" {
			reason = string(attr.Value)
		}
	}
	err := (*msAdapter).CallHangupWithReason(uuid, reason)
	if err != nil {
		return err
	}
	return nil
}

func ProcessHangupWithTiniyoReason(msAdapter *adapters.MediaServer, uuid string, reason string) error {
	err := (*msAdapter).CallHangupWithReason(uuid, reason)
	if err != nil {
		return err
	}
	return nil
}

func ProcessSyncHangup(msAdapter *adapters.MediaServer, uuid string, reason string) error {
	err := (*msAdapter).CallHangupWithSync(uuid, reason)
	if err != nil {
		return err
	}
	return nil
}