package tinixml

import (
	"fmt"
	"github.com/beevik/etree"
	"github.com/neoms/logger"
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
func ProcessConference(uuid string, element *etree.Element, authId string) string {
	logger.Logger.Debug("Creating Conference for " + uuid + "with name " + element.Text())
	confName := fmt.Sprintf("%s-%s@tiniyo", authId, element.Text())
	return confName
}
