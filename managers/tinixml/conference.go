package tinixml

import (
	"fmt"
	"github.com/beevik/etree"
	"net/url"
	"github.com/tiniyo/neoms/logger"
	"github.com/tiniyo/neoms/models"
)

/*
	<Response>
    	<Dial>
			<Conference>Test1234</Conference>
    	</Dial>
	</Response>

	Attributes:
	muted - Whether or not a caller can speak in a conference. Default is false.
	beep  - Whether or not a sound is played when callers leave or enter a conference. Default is true.
	startConferenceOnEnter - If a participant joins and startConferenceOnEnter is false,
							that participant will hear background music and stay muted until
							a participant with startConferenceOnEnter set.Default is true.
	endConferenceOnExit - If a participant with endConferenceOnExit set to true leaves a conference,
						the conference terminates and all participants drop out of the call. Default is false.

*/
func ProcessConference(callSid string, element *etree.Element, authId string) string {
	confName := fmt.Sprintf("%s-%s@tiniyo", authId, url.PathEscape(element.Text()))
	//check if conference already running
	//on which mediaserver its running
	//loopback to that mediaserver bridge conference
	//save attr to centralise system
	//query using conference name get attr also
	logger.UuidLog("Info", callSid, fmt.Sprintf("conference name is %s", confName))
	return confName
}

func ProcessConferenceAttr(data *models.CallRequest, child *etree.Element) {
	data.DialConferenceAttr.DialConferenceBeep = "true"
	for _, attr := range child.Attr {
		switch attr.Key {
		case "beep":
			if attr.Value == "false" || attr.Value == "onEnter" || attr.Value == "onExit" {
				data.DialConferenceAttr.DialConferenceBeep = attr.Value
			}
		default:
			logger.UuidLog("Err", data.ParentCallSid, fmt.Sprint("Attribute not supported - ", attr.Key))
		}
	}
}
