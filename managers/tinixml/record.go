package tinixml

import (
	"fmt"
	"github.com/beevik/etree"
	uuid4 "github.com/satori/go.uuid"
	"os"
	"github.com/tiniyo/neoms/adapters"
	"github.com/tiniyo/neoms/logger"
	"github.com/tiniyo/neoms/models"
	"strconv"
)

func ProcessRecord(msAdapter *adapters.MediaServer, data *models.CallRequest, child *etree.Element) error {
	if data.Status != "in-progress" {
		(*msAdapter).AnswerCall(data.CallSid)
	}
	var err error
	handleRecordAttribute(data, *child)
	recordingSid := uuid4.NewV4().String()
	recordMultiSet := fmt.Sprintf("^^:recording_sid=%s", recordingSid)
	if data.RecordFinishOnKey != "" {
		recordMultiSet = fmt.Sprintf(":playback_terminators=%s", data.RecordFinishOnKey)
	}
	if err = (*msAdapter).MultiSet(data.Sid, recordMultiSet); err != nil {
		return err
	}
	if data.RecordPlayBeep == "true" {
		if err = (*msAdapter).PlayBeep(data.Sid); err != nil {
			return err
		}
	}
	//now next is record_session with file name with record_stereo
	recordDir := "/call_recordings"
	recordFile := fmt.Sprintf("%s/%s-%s.mp3", recordDir, data.AccountSid, data.CallSid)
	if err = (*msAdapter).Record(data.Sid, recordFile, data.RecordMaxLength, data.RecordTimeout); err != nil {
		return err
	}
	return err
}

func handleRecordAttribute(data *models.CallRequest, child etree.Element) {
	callSid := data.Sid
	data.RecordingSource = "RecordVerb"
	data.RecordPlayBeep = "true"
	data.RecordMaxLength = "3600"
	data.RecordTimeout = "5"
	data.RecordFinishOnKey = "1234567890*#"
	logger.UuidLog("Info", callSid, fmt.Sprint("dial elements are", child))
	//recording attribute of file
	data.RecordMethod = "POST"
	for _, attr := range child.Attr {
		switch attr.Key {
		case "action":
			data.RecordAction = attr.Value
		case "method":
			data.RecordMethod = attr.Value
		case "timeout":
			if _, err := strconv.Atoi(attr.Value); err == nil {
				data.RecordTimeout = attr.Value
			}
		case "finishOnKey":
			data.RecordFinishOnKey = attr.Value
		case "maxLength":
			maxLen, err := strconv.Atoi(attr.Value)
			if err == nil && maxLen < 3600 {
				data.RecordMaxLength = attr.Value
			}
		case "playBeep":
			data.RecordPlayBeep = attr.Value
		case "trim":
			data.Trim = attr.Value
		case "recordingStatusCallback":
			data.RecordingStatusCallback = attr.Value
		case "recordingStatusCallbackMethod":
			data.RecordingStatusCallbackMethod = attr.Value
		case "recordingStatusCallbackEvent":
			data.RecordingStatusCallbackEvent = attr.Value
		case "storageUrl":
			data.RecordStorageUrl = attr.Value
		case "storageUrlMethod":
			data.RecordStorageUrlMethod = attr.Value
		case "transcribe":
			data.RecordTranscribe = attr.Value
		case "transcribeCallback":
			data.RecordTranscribeCallback = attr.Value
		default:
			logger.UuidLog("Err", callSid, fmt.Sprint("Attribute not supported - ", attr.Key))
		}
	}
	logger.UuidLog("Info", callSid, fmt.Sprint("dial variables are - ", data))
}

//We need to return  variable string set to be send together with originate
func ProcessRecordDialAttribute(data models.CallRequest) string {
	recordString := ""
	recordDir := "/call_recordings"

	//ensureRecordDir(recordDir)
	recordFile := fmt.Sprintf("%s/%s-%s.mp3", recordDir, data.AccountSid, data.Sid)
	switch data.Record {
	case "true", "record-from-answer":
		recordString = fmt.Sprintf("media_bug_answer_req=true,"+
			"api_on_answer_1='sched_api +1 none uuid_record %s start %s'", data.Sid,recordFile)
		if data.RecordingTrack == "inbound" {
			recordString = fmt.Sprintf("RECORD_READ_ONLY=true,media_bug_answer_req=true,"+
				"api_on_answer_1='sched_api +1 none uuid_record %s start %s'", data.Sid,recordFile)
		} else if data.RecordingTrack == "outbound" {
			recordString = fmt.Sprintf("RECORD_WRITE_ONLY=true,media_bug_answer_req=true,"+
				"api_on_answer_1='sched_api +1 none uuid_record %s start %s'", data.Sid,recordFile)
		}
	case "record-from-ringing":
		recordString = fmt.Sprintf("api_on_media_1='sched_api +1 none uuid_record %s start %s'", data.Sid,recordFile)
		if data.RecordingTrack == "inbound" {
			recordString = fmt.Sprintf("RECORD_READ_ONLY=true,api_on_media_1='sched_api +1 none  uuid_record %s start %s'",
				data.Sid,recordFile)
		} else if data.RecordingTrack == "outbound" {
			recordString = fmt.Sprintf("RECORD_WRITE_ONLY=true,api_on_media_1='sched_api +1 none  uuid_record %s start %s'",
				data.Sid,recordFile)
		}
	case "record-from-answer-dual":
		recordString = fmt.Sprintf("media_bug_answer_req=true,RECORD_STEREO=true,"+
			"api_on_answer_1='sched_api +1 none uuid_record %s start %s'", data.Sid,recordFile)
		if data.RecordingTrack == "inbound" {
			recordString = fmt.Sprintf("RECORD_READ_ONLY=true,"+
				"api_on_answer_1='sched_api +1 none uuid_record %s start %s'", data.Sid,recordFile)
		} else if data.RecordingTrack == "outbound" {
			recordString = fmt.Sprintf("RECORD_WRITE_ONLY=true,"+
				"api_on_answer_1='sched_api +1 none  uuid_record %s start %s'", data.Sid,recordFile)
		}
	case "record-from-ringing-dual":
		recordString = fmt.Sprintf("RECORD_STEREO=true,api_on_media_1='sched_api +1 none uuid_record %s start %s'",
			data.Sid,recordFile)
		if data.RecordingTrack == "inbound" {
			recordString = fmt.Sprintf("RECORD_READ_ONLY=true,api_on_media_1='sched_api +1 none uuid_record %s start %s'",
				data.Sid,recordFile)
		} else if data.RecordingTrack == "outbound" {
			recordString = fmt.Sprintf("RECORD_WRITE_ONLY=true,api_on_media_1='sched_api +1 none uuid_record %s start %s'",
				data.Sid,recordFile)
		}
	default:
	}

	return recordString
}

func ensureRecordDir(dirName string) error {
	err := os.Mkdir(dirName, 0755)
	if err == nil || os.IsExist(err) {
		return nil
	} else {
		return err
	}
}
