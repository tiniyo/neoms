package tinixml

import (
	"fmt"
	"github.com/beevik/etree"
	"github.com/tiniyo/neoms/adapters"
	"github.com/tiniyo/neoms/constant"
	"github.com/tiniyo/neoms/logger"
	"github.com/tiniyo/neoms/models"
	"strings"
	"time"
)

func ProcessRedirect(callSid string,msAdapter *adapters.MediaServer, data models.CallRequest,element *etree.Element) (error,string, string) {
	redirectUrl := element.Text()
	redirectMethod := "POST"
	if redirectUrl == "" {
		logger.UuidLog("Info",callSid,fmt.Sprintf("Received Empty Redirect url, Processing the current xml sequence - %s", redirectUrl))
		return constant.ErrEmptyRedirect, "",""
	}
	logger.UuidLog("Info",callSid, fmt.Sprintf("Received Empty Redirect url, Processing the current xml sequence - %s", redirectUrl))
	for _, attr := range element.Attr {
		if attr.Key == "method" {
			redirectMethod = strings.ToUpper(attr.Value)
		}
	}

	_ = (*msAdapter).PlayMediaFile(data.CallSid, "silence_stream://500", "1")

	for {
		if emptyUuidQueue, err := (*msAdapter).UuidQueueCount(callSid); err != nil {
			break
		} else if !emptyUuidQueue {
			time.Sleep(10 * time.Millisecond)
		} else {
			break
		}
	}

	return nil, redirectUrl, redirectMethod
}

