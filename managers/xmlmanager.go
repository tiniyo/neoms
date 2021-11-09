package managers

import (
	"fmt"
	"net/url"
	"github.com/tiniyo/neoms/adapters"
	"github.com/tiniyo/neoms/constant"
	"strings"
	"time"

	"github.com/beevik/etree"
	"github.com/tiniyo/neoms/helper"
	"github.com/tiniyo/neoms/logger"
	. "github.com/tiniyo/neoms/managers/tinixml"
	"github.com/tiniyo/neoms/models"
)

type XmlManagerInterface interface {
	ParseTinyXml(data models.CallRequest, resp []byte) error
	ProcessXmlResponse(data models.CallRequest)
	handleXmlUrl(data models.CallRequest) error
	requestForXml(xmlUrl string, xmlMethod string, callSid string, dataMap map[string]interface{}) (bool, []byte)
}

type XmlManager struct {
	callState  adapters.CallStateAdapter
}

func NewXmlManager(callStateAdapter adapters.CallStateAdapter) XmlManagerInterface {
	return XmlManager{
		callState:  callStateAdapter,
	}
}

/*
	for inbound call - ProcessXmlResponse will call when call_park event received
	for outbound-api call - ProcessXmlResponse will get call when call_answer event received
*/

func (xmlMgr XmlManager) ParseTinyXml(data models.CallRequest, resp []byte) error {
	var err error
	nextElement := true
	intercept := false
	isRedirect := false
	doc := etree.NewDocument()
	uuid := data.CallSid
	var root = new(etree.Element)

	if uuid == "" {
		uuid = data.Sid
		data.CallSid = uuid
	}

	//before processing first stop other processing if any other xml is on going on same call
	// we might need to check on redis if any xml is getting processed in this call
	if err = MsAdapter.BreakAllUuid(data.CallSid); err != nil {
		logger.UuidLog("Err", uuid, fmt.Sprintf("sending uuid break command failed, live with it - %#v", err))
	}

	logger.UuidLog("Info", uuid, fmt.Sprintf("xml parsing started"))
	if resp == nil {
		logger.UuidLog("Err", uuid, fmt.Sprintf("xml parsing stopped,hangup call"))
	} else if err = doc.ReadFromBytes(resp); err != nil {
		logger.UuidLog("Err", uuid, fmt.Sprintf("xml parsing stopped,hangup call"))
	} else if doc == nil {
		logger.UuidLog("Err", uuid, fmt.Sprintf("xml parsing stopped,hangup call"))
	} else if root = doc.SelectElement("Response"); root != nil {
		xmlChildes := root.ChildElements()

		for _, xmlChild := range xmlChildes {
			switch xmlChild.Tag {
			case "Reject":
				nextElement = false
				err = ProcessReject(&MsAdapter, uuid, xmlChild)
			case "Play":
				err = ProcessPlay(&MsAdapter, data, xmlChild)
			case "Dial":
				if data.DialNumberUrl == "" {
					nextElement, intercept, err = ProcessDial(&MsAdapter, data, xmlChild)
				}
				// we need to wait for dial to success or fail here
				
				if intercept {
					//this is special condition we are going to wait here xml to finish at other leg
					for {
						time.Sleep(1 * time.Second)
						parentCallSid := fmt.Sprintf("intercept:%s", data.CallSid)
						if val, err := xmlMgr.callState.KeyExist(parentCallSid); err != nil || !val {
							logger.UuidLog("Err", uuid, fmt.Sprintf("key does not set for intercept - wait"))
						} else {
							logger.UuidLog("Err", uuid, fmt.Sprintf("key set for intercept - processing with next element"))
							intercept = false
							break
						}
					}
				}
			case "Stream":
			case "Siprec":
			case "Refer":
			case "Record":
				nextElement = false
				if err = ProcessRecord(&MsAdapter, &data, xmlChild); err == nil {
					//get the json request of call request
					if dataByte, err := json.Marshal(data); err == nil {
						_ = xmlMgr.callState.Set(uuid, dataByte)
					}
				}
			case "Pay":
			case "Leave":
			case "Gather":
				nextElement, err = ProcessGather(&MsAdapter, &data, xmlChild)
				if err != nil && err.Error() == "TIMEOUT" {
					return constant.ErrGatherTimeout
				}
			case "Autopilot":
			case "Enqueue":
			case "Speak", "Say":
				err = ProcessSpeak(&MsAdapter, data, xmlChild)
			case "Redirect":
				if err, redirectUrl, redirectMethod := ProcessRedirect(uuid, &MsAdapter, data, xmlChild); err == nil {
					nextElement = false
					isRedirect = true
					statusCallbackKey := fmt.Sprintf("statusCallback:%s", uuid)
					val, err := xmlMgr.callState.Get(statusCallbackKey)
					if err != nil {
						logger.UuidLog("Err", uuid, fmt.Sprintf("redirect url - unmarshal failed %s", err.Error()))
					} else if err = json.Unmarshal(val, &data); err != nil {
						logger.UuidLog("Err", uuid, fmt.Sprintf("redirect url - unmarshal failed %s", err.Error()))
					} else if data.HangupTime == "" {
						data.Url = redirectUrl
						data.Method = redirectMethod
						_ = xmlMgr.handleXmlUrl(data)
					}
				}
			case "Hangup":
				nextElement = false
				if err = ProcessHangup(&MsAdapter, uuid, xmlChild); err != nil {
					// handle error here
				}
				break
			case "Pause":
				ProcessPause(data.Sid, xmlChild)
			default:
				logger.UuidLog("Info", uuid, fmt.Sprintf("xml child tag is %s", xmlChild.Tag))
			}

			if !nextElement {
				break
			}
		}
	}

	if isRedirect {
		return nil
	} else if data.DialNumberUrl != "" {
		logger.UuidLog("Info", uuid, fmt.Sprintf("We are going to bridge parent and child call here"))
		if err = MsAdapter.CallIntercept(data.CallSid, data.ParentCallSid); err != nil {
			if err = ProcessSyncHangup(&MsAdapter, data.CallSid, "XML_CallFlow_Complete"); err != nil {
				logger.UuidLog("Err", uuid, fmt.Sprintf("sending call hangup event failed - %s", err.Error()))
				return err
			}
		}
	}

	logger.UuidLog("Info", uuid, fmt.Sprintf("sending synchronous call hangup"))
	if err = ProcessSyncHangup(&MsAdapter, data.CallSid, "XML_CallFlow_Complete"); err != nil {
		logger.UuidLog("Err", uuid, fmt.Sprintf("call hangup failed - %s", err.Error()))
		return err
	}
	return nil
}

func (xmlMgr XmlManager) ProcessXmlResponse(data models.CallRequest) {
	uuid := data.Sid
	logger.UuidLog("Info", uuid, fmt.Sprintf("processing xml response"))
	if data.Speak != "" {
		logger.Logger.Info(data.Speak)
		_ = ProcessSpeakText(&MsAdapter, uuid, data.Speak)
		_ = ProcessHangupWithTiniyoReason(&MsAdapter, uuid, "NORMAL_CLEARING")
	} else if data.Play != "" {
		_ = ProcessPlayFile(&MsAdapter, uuid, data.Play)
		_ = ProcessHangupWithTiniyoReason(&MsAdapter, uuid, "NORMAL_CLEARING")
	} else {
		logger.UuidLog("Info", uuid, fmt.Sprintf("url found getting xml"))
		_ = xmlMgr.handleXmlUrl(data)
	}
}

func (xmlMgr XmlManager) handleXmlUrl(data models.CallRequest) error {
	xmlUrl := data.Url
	callSid := data.Sid
	xmlMethod := strings.ToUpper(data.Method)
	dataMap := make(map[string]interface{})

	if data.SipTrunk == "true" {
		xmlData := []byte(`<?xml version="1.0" encoding="UTF-8" ?>
<Response>
    <Dial answerOnBridge="true" timeout="20">
        <Number>` + data.To + `</Number>
    </Dial>
</Response>`)
		return xmlMgr.ParseTinyXml(data, xmlData)
	}

	if data.TinyML != "" {
		logger.UuidLog("Info", callSid, fmt.Sprintf("handleXmlUrl tinyml - %s", data.TinyML))
		if tinyMl, err := url.QueryUnescape(data.TinyML); err == nil {
			return xmlMgr.ParseTinyXml(data, []byte(tinyMl))
		}
		logger.UuidLog("Err", callSid, fmt.Sprint("handleXmlUrl tinyml parsing error - "))
		return ProcessHangupWithTiniyoReason(&MsAdapter, callSid, "UNALLOCATED_NUMBER")
	}

	logger.UuidLog("Info", callSid, fmt.Sprintf("handleXmlUrl url - %s, method - %s", xmlUrl, xmlMethod))
	if byteData, err := json.Marshal(data.Callback); err == nil {
		if err = json.Unmarshal(byteData, &dataMap); err != nil {
			logger.UuidLog("Err", callSid, fmt.Sprint("send url request failed - ", err))
			return ProcessHangupWithTiniyoReason(&MsAdapter, callSid, "UNALLOCATED_NUMBER")
		}
	} else {
		logger.UuidLog("Err", callSid, fmt.Sprint("send url request failed - ", err))
		return ProcessHangupWithTiniyoReason(&MsAdapter, callSid, "UNALLOCATED_NUMBER")
	}

	if status, respBody := xmlMgr.requestForXml(xmlUrl, xmlMethod, callSid, dataMap); status {
		return xmlMgr.ParseTinyXml(data, respBody)
	}

	xmlUrl = data.FallbackUrl
	xmlMethod = data.FallbackMethod

	if status, respBody := xmlMgr.requestForXml(xmlUrl, xmlMethod, callSid, dataMap); status {
		return xmlMgr.ParseTinyXml(data, respBody)
	}
	return ProcessHangupWithTiniyoReason(&MsAdapter, callSid, "Failed_To_Get_XML")

}

func (xmlMgr XmlManager) requestForXml(xmlUrl string, xmlMethod string, callSid string, dataMap map[string]interface{}) (bool, []byte) {
	if xmlUrl == "" {
		return false, nil
	}
	if xmlMethod == "" {
		xmlMethod = "POST"
	}
	switch xmlMethod {
	case "GET", "get", "Get":
		statusCode, respBody, err := helper.Get(callSid, dataMap, xmlUrl)
		if err != nil || statusCode != 200 || respBody == nil {
			logger.UuidLog("Err", callSid, fmt.Sprintf("Error while getting the GET XML %v", err))
			return false, nil
		}
		logger.UuidLog("Info", callSid, fmt.Sprintf(" GET XML success %v", statusCode))
		/*if len(respBody) > 0 {
			isValid, errstr := utils.ValidateXML(respBody)
			if isValid == false {
				logger.UuidLog("Err", callSid, fmt.Sprintf("Error while parsing the XML %v", errstr))
				return false, nil
			}
		}*/
		return true, respBody
	case "POST", "post", "Post":
		statusCode, respBody, err := helper.Post(callSid, dataMap, xmlUrl)
		if err != nil || statusCode != 200 || respBody == nil {
			logger.UuidLog("Err", callSid, fmt.Sprintf("Error while getting the POST XML %v", err))
			return false, nil
		}
		logger.UuidLog("Info", callSid, fmt.Sprintf(" POST XML success %v", statusCode))
		/*if len(respBody) > 0 {
			isValid, errstr := utils.ValidateXML(respBody)
			if isValid == false {
				logger.UuidLog("Err", callSid, fmt.Sprintf("Error while parsing the XML %v", errstr))
				return false, nil
			}
		}*/
		return true, respBody
	default:
		logger.UuidLog("Info", callSid, fmt.Sprintf("Unknown Method url - %s, method - %s", xmlUrl, xmlMethod))
		_ = ProcessHangupWithTiniyoReason(&MsAdapter, callSid, "Failed_To_Get_XML")
	}
	return false, nil
}
