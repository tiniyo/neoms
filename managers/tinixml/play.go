package tinixml

import (
	"fmt"
	"github.com/beevik/etree"
	"github.com/tiniyo/neoms/adapters"
	"github.com/tiniyo/neoms/logger"
	"github.com/tiniyo/neoms/models"
	"strconv"
)

/*
	playing from xml
	<Response>
    	<Play>https://s3.amazonaws.com/tiniyo/Trumpet.mp3</Play>
	</Response>
*/
func ProcessPlay(msAdapter *adapters.MediaServer, data models.CallRequest, element *etree.Element) error {
	callSid := data.CallSid
	if data.Status != "in-progress"{
		(*msAdapter).AnswerCall(data.CallSid)
	}

	loopCount := 1
	for _, attr := range element.Attr {
		logger.Logger.Debug("ATTR: %s=%s\n", attr.Key, attr.Value)
		if attr.Key == "loop" {
			loopCount, _ = strconv.Atoi(attr.Value)
			if loopCount == 0 {
				loopCount = 1
			}
		}
	}
	strLoopCount := strconv.Itoa(loopCount)
	if err := (*msAdapter).PlayMediaFile(callSid, element.Text(), strLoopCount);err != nil {
		logger.UuidLog("Err", callSid, fmt.Sprint("error while playing media file", err))
		return err
	}
	return nil
}

/*
	playing from rest
	{
		"to":"your_destination",
		"from":"your_callerId",
		"play":"url for file"
	}
*/
func ProcessPlayFile(msAdapter *adapters.MediaServer, uuid string, fileUrl string) error {
	loopCount := 3
	strLoopCount := strconv.Itoa(loopCount)
	err := (*msAdapter).PlayMediaFile(uuid, fileUrl, strLoopCount)
	if err != nil {
		return err
	}
	return nil
}
