package mediaserver

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/fiorix/go-eventsocket/eventsocket"
	"github.com/neoms/adapters"
	"github.com/neoms/config"
	"github.com/neoms/logger"
)

/*
	MsFreeSWITCHGiz have 2 connection to FreeSWITCH
	one for getting events and another for sending commands
*/
type MsFreeSWITCHGiz struct {
	c  *eventsocket.Connection
	c1 *eventsocket.Connection
}

const (
	subEventList = "events text CHANNEL_PARK " +
		"CHANNEL_DESTROY CHANNEL_EXECUTE_COMPLETE " +
		"CHANNEL_HANGUP CHANNEL_UNBRIDGE " +
		"CHANNEL_BRIDGE CHANNEL_ANSWER " +
		"RECORD_START RECORD_STOP CHANNEL_ORIGINATE " +
		"CHANNEL_PROGRESS CHANNEL_PROGRESS_MEDIA " +
		"CHANNEL_HANGUP_COMPLETE DTMF SESSION_HEARTBEAT " +
		"MESSAGE CUSTOM conference::maintenance"
	freeSwitchReconnectTime = 5
)

func connect(cb adapters.MediaServerCallbacker) (*eventsocket.Connection, error) {
	dialTo := config.Config.Fs.FsHost + ":" + config.Config.Fs.FsPort
	logger.Logger.Debug("FreeSWITCH Dial To : ", dialTo)
	c, err := eventsocket.Dial(dialTo, config.Config.Fs.FsPassword)
	if err != nil {
		logger.Logger.Error("FreeSWITCH Connection Failed - ", err)
		if cb != nil {
			cb.CallBackMediaServerStatus(0)
		}
		return c, err
	}
	if cb != nil {
		_ = cb.CallBackMediaServerStatus(1)
	}
	return c, nil
}

func reconnect(cb adapters.MediaServerCallbacker) (*eventsocket.Connection, error) {
	logger.Logger.Info("Trying to reconnect to FreeSWITCH")
	for {
		c, err := connect(cb)
		if err == nil {
			logger.Logger.Error("FreeSWITCH Connection Success - ", err)
			return c, err
		}
		time.Sleep(time.Second * freeSwitchReconnectTime)
		logger.Logger.Info("Trying to reconnect to FreeSWITCH in 5 Seconds")
	}
}

func (fs *MsFreeSWITCHGiz) InitializeCallbackMediaServers(cb adapters.MediaServerCallbacker) error {
	/* Spawn new thread as it will not stop execution during initialization */
	go func() {
		go fs.initializeEventCallbackConnection(cb)
		var err error
		fs.c, err = reconnect(cb)
		if err != nil {
			logger.Logger.Error("Initialize Mediaserver Failed ", err)
		}
	}()
	return nil
}

func (fs *MsFreeSWITCHGiz) initializeEventCallbackConnection(cb adapters.MediaServerCallbacker) error {
	if cb == nil {
		return nil
	}
	/* Spawn new thread as it will not stop execution during initialization */
	go func() {
		var err error
		fs.c1, err = reconnect(cb)
		if err != nil {
			logger.Logger.Error("Initialize Mediaserver Failed ", err)
		}
		fs.c1.Send(subEventList)
		logger.Logger.Info("FreeSWITCH Connected, subscribed event list ", subEventList)
		go fs.run(cb)
	}()
	return nil
}

func handleFreeSWITCHEvent(ev *eventsocket.Event, cb adapters.MediaServerCallbacker) {
	evName := ev.Get("Event-Name")
	callSid := ev.Get("Variable_call_sid")
	if callSid == ""{
		callSid = ev.Get("Unique-ID")
	}
	parentCallSid := ev.Get("Variable_parent_call_sid")
	logger.Logger.Debug("UniqueuID : ", callSid, " Event Name : ", evName)
	evHeader, _ := json.Marshal(ev.Header)
	switch evName {
	case "CHANNEL_PARK":
		cb.CallBackPark(callSid, evHeader)
	case "CHANNEL_DESTROY":
		cb.CallBackDestroy(callSid)
	case "CHANNEL_EXECUTE_COMPLETE":
		cb.CallBackExecuteComplete(callSid)
	case "CHANNEL_HANGUP":
		cb.CallBackHangup(callSid)
	case "CHANNEL_HANGUP_COMPLETE":
		cb.CallBackHangupComplete(callSid, evHeader)
	case "CHANNEL_UNBRIDGE":
		cb.CallBackUnBridged(callSid)
	case "CHANNEL_BRIDGE":
		cb.CallBackBridged(callSid)
	case "CHANNEL_ANSWER":
		cb.CallBackAnswered(callSid, evHeader)
	case "CHANNEL_PROGRESS":
		cb.CallBackProgress(callSid)
	case "CHANNEL_PROGRESS_MEDIA":
		cb.CallBackProgressMedia(callSid, evHeader)
	case "SESSION_HEARTBEAT":
		cb.CallBackSessionHeartBeat(parentCallSid, callSid)
	case "RECORD_START":
		cb.CallBackRecordingStart(callSid, evHeader)
	case "RECORD_STOP":
		cb.CallBackRecordingStop(callSid, evHeader)
	case "MESSAGE":
		cb.CallBackMessage(callSid)
	case "CUSTOM":
		cb.CallBackCustom(callSid)
	case "CHANNEL_ORIGINATE":
		cb.CallBackOriginate(callSid, evHeader)
	case "DTMF":
		cb.CallBackDTMFDetected(callSid, evHeader)
	default:
		logger.Logger.Debug("UniqueID : ", callSid, " Event Name : ", evName)
	}
}

func (fs *MsFreeSWITCHGiz) run(cb adapters.MediaServerCallbacker) {
	for {
		ev, err := fs.c1.ReadEvent()
		if err != nil {
			logger.Logger.Error("FreeSWITCH ESL Read event failed ", err)
			logger.Logger.Info("Reconnect to the FreeSWITCH again ")
			go fs.InitializeCallbackMediaServers(cb)
			break
		}
		go handleFreeSWITCHEvent(ev, cb)
	}
}

func (fs MsFreeSWITCHGiz) EnableSessionHeartBeat(uuid, interval string) error {
	logger.Logger.Debug("EnableSessionHeartBeat : ", uuid)
	type MSG map[string]string
	ev, err := fs.c.SendMsg(eventsocket.MSG{
		"call-command":     "execute",
		"execute-app-name": "enable_heartbeat",
		"execute-app-arg":  interval,
		"event-lock":       "false"}, uuid, "")
	if err != nil {
		logger.Logger.Debug("EnableSessionHeartBeat Event : ", ev, " Error:", err, uuid)
		return err
	}
	return nil
}

func (fs MsFreeSWITCHGiz) PreAnswerCall(uuid string) error {
	logger.Logger.Debug("Pre Answer the call with : ", uuid)
	type MSG map[string]string
	_, err := fs.c.SendMsg(eventsocket.MSG{
		"call-command":     "execute",
		"execute-app-name": "pre_answer",
		"event-lock":       "true"}, uuid, "")
	if err != nil {
		logger.Logger.Error("Error while Pre Answer Call - ", err)
		return err
	}
	/*
		This to make sure command should be relayed after ack is received
		ack is expected with in 1 second else issues will be there
	*/
	time.Sleep(1*time.Second)
	return nil
}

func (fs MsFreeSWITCHGiz) AnswerCall(uuid string) error {
	logger.Logger.Debug("Answering the call with : ", uuid)
	type MSG map[string]string
	_, err := fs.c.SendMsg(eventsocket.MSG{
		"call-command":     "execute",
		"execute-app-name": "answer",
		"event-lock":       "true"}, uuid, "")
	if err != nil {
		logger.Logger.Error("Error while AnswerCall - ", err)
		return err
	}
	/*
		This to make sure command should be relayed after ack is received
		ack is expected with in 1 second else issues will be there
	 */
	time.Sleep(1*time.Second)
	return nil
}

func (fs MsFreeSWITCHGiz) PlayMediaFile(uuid string, fileUrl string, loopCount string) error {
	logger.Logger.Debug("PlayMediaFile : ", uuid, " FileURL: ", fileUrl)
	type MSG map[string]string
	_, err := fs.c.SendMsg(eventsocket.MSG{
		"call-command":     "execute",
		"execute-app-name": "playback",
		"execute-app-arg":  fileUrl,
		"loops":            loopCount,
		"event-lock":       "true"}, uuid, "")
	if err != nil {
		logger.Logger.Error("Error while playing media file - ", err)
		return err
	}
	return nil
}

func (fs MsFreeSWITCHGiz) PlayBeep(uuid string) error {
	type MSG map[string]string
	_, err := fs.c.SendMsg(eventsocket.MSG{
		"call-command":     "execute",
		"execute-app-name": "playback",
		"execute-app-arg":  "tone_stream://L=1;%(1850,1750,1000)",
		"event-lock":       "true"}, uuid, "")
	if err != nil {
		logger.Logger.Error("Error while playing beap  - ", err)
		return err
	}
	return nil
}

func (fs MsFreeSWITCHGiz) Speak(uuid string, voiceId, text string) error {
	args := fmt.Sprintf("polly|%s|%s", voiceId, text)
	type MSG map[string]string
	_, err := fs.c.SendMsg(eventsocket.MSG{
		"call-command":     "execute",
		"execute-app-name": "speak",
		"execute-app-arg":  args,
		"event-lock":       "true"}, uuid, "")
	if err != nil {
		logger.Logger.Error("Error while playing media file - ", err)
		return err
	}
	return nil
}

func (fs MsFreeSWITCHGiz) CallNewOutbound(cmd string) error {
	_, err := fs.c.Send(cmd)
	if err != nil {
		logger.Logger.Error("Error while executing the new outbound command", err)
		return err
	}
	return nil
}


func (fs MsFreeSWITCHGiz) CallHangup(uuid string) error {
	logger.Logger.Debug("Hangup : ", uuid)
	_, err := fs.c.SendMsg(eventsocket.MSG{
		"call-command":     "execute",
		"execute-app-name": "hangup",
		"event-lock":       "true"}, uuid, "")
	if err != nil {
		logger.Logger.Error("Error while executing the hangup command", err)
		return err
	}
	return nil
}

func (fs MsFreeSWITCHGiz) CallHangupWithSync(uuid string, reason string) error {
	logger.Logger.Debug("Hangup : ", uuid)
	ev, err := fs.c.SendMsg(eventsocket.MSG{
		"call-command":     "execute",
		"execute-app-name": "hangup",
		"execute-app-arg":  reason,
		"event-lock":       "true"}, uuid, "")
	logger.Logger.Debug("Destination Event : ", ev, " Error:", err)
	return nil
}

func (fs MsFreeSWITCHGiz) CallHangupWithReason(uuid string, reason string) error {
	logger.Logger.Debug("Hangup : ", uuid)
	ev, err := fs.c.SendMsg(eventsocket.MSG{
		"call-command":     "execute",
		"execute-app-name": "hangup",
		"execute-app-arg":  reason,
		"event-lock":       "false"}, uuid, "")
	logger.Logger.Debug("Destination Event : ", ev, " Error:", err)
	return nil
}

func (fs MsFreeSWITCHGiz) CallTransfer() error {
	return nil
}

func (fs MsFreeSWITCHGiz) CallSendDTMF(uuid string, dtmf string) error {
	logger.Logger.Debug("SendDTMF  : ", uuid)
	ev, err := fs.c.SendMsg(eventsocket.MSG{
		"call-command":     "execute",
		"execute-app-name": "send_dtmf",
		"execute-app-arg":  dtmf,
		"event-lock":       "true"}, uuid, "")
	logger.Logger.Debug("Destination Event : ", ev, " Error:", err)
	if err != nil {
		return err
	}
	return nil
}

func (fs MsFreeSWITCHGiz) CallReceiveDTMF(uuid string) error {
	logger.Logger.Debug("ReceiveDTMF  : ", uuid)
	ev, err := fs.c.Send(fmt.Sprintf("bgapi uuid_recv_dtmf %s %s", uuid))
	if err != nil {
		return err
	}
	logger.Logger.Debug("Destination Event : ", ev, " Error:", err)
	return nil
}


func (fs MsFreeSWITCHGiz) BreakAllUuid(uuid string) error  {
	logger.Logger.Debug("Break All Command to Uuid  : ", uuid)
	ev, err := fs.c.Send(fmt.Sprintf("bgapi uuid_break %s all", uuid))
	if err != nil {
		return err
	}
	logger.Logger.Debug("Destination Event : ", ev, " Error:", err)
	return nil
}


func (fs MsFreeSWITCHGiz) UuidQueueCount(uuid string) (bool,error)  {
	ev, err := fs.c.Send(fmt.Sprintf("api uuid_queue_count %s", uuid))
	if err != nil {
		return false, err
	}
	if ev != nil && strings.HasPrefix(ev.Body,"+OK 0") {
		return true, nil
	}
	return false, nil
}

func (fs MsFreeSWITCHGiz) CallBridge(uuid string, bridgeArgs string) error {
	logger.Logger.Debug("Bridge : ", uuid, " And ", bridgeArgs)
	ev, err := fs.c.SendMsg(eventsocket.MSG{
		"call-command":     "execute",
		"execute-app-name": "bridge",
		"execute-app-arg":  bridgeArgs,
		"event-lock":       "true"}, uuid, "")
	logger.Logger.Debug("Destination Event : ", ev, " Error:", err)
	if err != nil {
		return err
	}
	return nil
}

func (fs MsFreeSWITCHGiz) CallIntercept(uuid string, bridgeArgs string) error {
	logger.Logger.Debug("Bridge : ", uuid, " And ", bridgeArgs)
	ev, err := fs.c.SendMsg(eventsocket.MSG{
		"call-command":     "execute",
		"execute-app-name": "intercept",
		"execute-app-arg":  bridgeArgs,
		"event-lock":       "true"}, uuid, "")
	logger.Logger.Debug("Destination Event : ", ev, " Error:", err)
	if err != nil {
		return err
	}
	return nil
}

func (fs MsFreeSWITCHGiz) SetRecordStereo(uuid string) error {
	logger.Logger.Debug("SetRecordStereo : ", uuid)
	ev, err := fs.c.SendMsg(eventsocket.MSG{
		"call-command":     "execute",
		"execute-app-name": "set",
		"execute-app-arg":  "RECORD_STEREO=true",
		"event-lock":       "true"}, uuid, "")
	logger.Logger.Debug("Destination Event : ", ev, " Error:", err)
	return err
}

func (fs MsFreeSWITCHGiz) Set(uuid, value string) error {
	logger.Logger.Debug("Set Variable : ", uuid)
	ev, err := fs.c.SendMsg(eventsocket.MSG{
		"call-command":     "execute",
		"execute-app-name": "set",
		"execute-app-arg":  value,
		"event-lock":       "false"}, uuid, "")
	logger.Logger.Debug("Destination Event : ", ev, " Error:", err)
	if err != nil {
		return err
	}
	return nil
}

func (fs MsFreeSWITCHGiz) MultiSet(uuid, value string) error {
	logger.Logger.Debug("Set Variable : ", uuid)
	ev, err := fs.c.SendMsg(eventsocket.MSG{
		"call-command":     "execute",
		"execute-app-name": "multiset",
		"execute-app-arg":  value,
		"event-lock":       "false"}, uuid, "")
	logger.Logger.Debug("Destination Event : ", ev, " Error:", err)
	if err != nil {
		return err
	}
	return nil
}

func (fs MsFreeSWITCHGiz) CallRecord(uuid string, recordFile string) error {
	ev, err := fs.c.SendMsg(eventsocket.MSG{
		"call-command":     "execute",
		"execute-app-name": "record_session",
		"execute-app-arg":  recordFile,
		"event-lock":       "true"}, uuid, "")
	logger.Logger.Debug("Destination Event : ", ev, " Error:", err)
	if err != nil {
		return err
	}
	return nil
}

func (fs MsFreeSWITCHGiz) Record(uuid string, recordFile string,maxDuration string, silenceSeconds string) error {
	recordString := fmt.Sprintf("%s %s 40 %s",recordFile,maxDuration,silenceSeconds)
	if silenceSeconds == "0"{
		recordString = fmt.Sprintf("'%s %s'",recordFile,maxDuration)
	}
	ev, err := fs.c.SendMsg(eventsocket.MSG{
		"call-command":     "execute",
		"execute-app-name": "record",
		"execute-app-arg":  recordString,
		"event-lock":       "true"}, uuid, "")
	logger.Logger.Debug("Destination Event : ", ev, " Error:", err)
	if err != nil {
		return err
	}
	return nil
}


//Conference
func (fs MsFreeSWITCHGiz) ConfCreate(uuid, conferenceName string) error {
	logger.Logger.Debug("Conference : ", uuid, " And ", conferenceName)
	ev, err := fs.c.SendMsg(eventsocket.MSG{
		"call-command":     "execute",
		"execute-app-name": "conference",
		"execute-app-arg":  conferenceName,
		"event-lock":       "true"}, uuid, "")
	logger.Logger.Debug("Destination Event : ", ev, " Error:", err)
	if err != nil {
		return err
	}
	return nil
}

//ConferenceBridge
func (fs MsFreeSWITCHGiz) ConfBridge(uuid, bridgeArgs string) error {
	logger.Logger.Debug("Bridge : ", uuid, " And ", bridgeArgs)
	ev, err := fs.c.SendMsg(eventsocket.MSG{
		"call-command":     "execute",
		"execute-app-name": "conference",
		"execute-app-arg":  bridgeArgs,
		"event-lock":       "true"}, uuid, "")
	logger.Logger.Debug("Destination Event : ", ev, " Error:", err)
	if err != nil {
		return err
	}
	return nil
}

//ConferenceBridge
func (fs MsFreeSWITCHGiz) ConfSetAutoCall(uuid, bridgeArgs string) error {
	logger.Logger.Debug("ConfSetAutoCall : ", uuid, " And ", bridgeArgs)
	ev, err := fs.c.SendMsg(eventsocket.MSG{
		"call-command":     "execute",
		"execute-app-name": "conference_set_auto_outcall",
		"execute-app-arg":  bridgeArgs,
		"event-lock":       "true"}, uuid, "")
	logger.Logger.Debug("Destination Event : ", ev, " Error:", err)
	if err != nil {
		return err
	}
	return nil
}

func (fs MsFreeSWITCHGiz) ConfAddMember() error {
	return nil
}

func (fs MsFreeSWITCHGiz) ConfRemoveMember() error {
	return nil
}
