package mediaserver

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/fiorix/go-eventsocket/eventsocket"
	"github.com/tiniyo/neoms/adapters"
	"github.com/tiniyo/neoms/config"
	"github.com/tiniyo/neoms/logger"
)

/*
	MsFreeSWITCHGiz have 2 connection to FreeSWITCH
	one for getting events and another for sending commands
*/
type MsFreeSWITCHGiz struct {
	fsConn *eventsocket.Connection
	eventReceiver *eventsocket.Connection
}

const (
	subEventList = "events text CHANNEL_PARK " +
		"CHANNEL_ANSWER " +
		"RECORD_START RECORD_STOP CHANNEL_ORIGINATE " +
		"CHANNEL_PROGRESS_MEDIA " +
		"CHANNEL_HANGUP_COMPLETE DTMF SESSION_HEARTBEAT"
	//	"CUSTOM conference::maintenance"
	freeSwitchReconnectTime = 5
)

func connect(cb adapters.MediaServerCallbacker) (*eventsocket.Connection, error) {
	dialTo := config.Config.Fs.FsHost + ":" + config.Config.Fs.FsPort
	logger.Logger.Debug("FreeSWITCH Dial To : ", dialTo)
	c, err := eventsocket.Dial(dialTo, config.Config.Fs.FsPassword)
	if err != nil {
		logger.Logger.Error("FreeSWITCH Connection Failed - ", err)
		if cb != nil {
			err = cb.CallBackMediaServerStatus(0)
			if err != nil {
				return nil, err
			}
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
			return c, nil
		}
		time.Sleep(time.Second * freeSwitchReconnectTime)
		logger.Logger.Info("Trying to reconnect to FreeSWITCH in 5 Seconds")
	}
}

func (fs *MsFreeSWITCHGiz) InitializeCallbackMediaServers(cb adapters.MediaServerCallbacker) error {
	/* Spawn new thread as it will not stop execution during initialization */
	go func() {
		//initialize callback connection
		go func() {
			err := fs.initializeEventCallbackConnection(cb)
			if err != nil {
				logger.Logger.Error("Initialize MediaServer Callback Connection Failed- ", err)
			} else {
				logger.Logger.Error("Initialize MediaServer Callback Connection Success- ", err)
			}
		}()

		//initialize callback connection
		var err error
		fs.fsConn, err = reconnect(cb)
		if err != nil {
			logger.Logger.Error("Initialize MediaServer Failed ", err)
		} else {
			logger.Logger.Error("Initialize MediaServer success ", err)
		}

	}()
	return nil
}

func (fs *MsFreeSWITCHGiz) initializeEventCallbackConnection(cb adapters.MediaServerCallbacker) error {
	if cb == nil {
		logger.Logger.Error("Initialize MediaServer Callback Connection Failed- callback is nil")
		return nil
	}
	/* Spawn new thread as it will not stop execution during initialization */
	go func() {
		var err error
		//infinite loop for connecting to freeswitch and subscribe to events
		for {
			fs.eventReceiver, err = reconnect(cb)
			if err != nil {
				logger.Logger.Error("Initialize MediaServer Failed ", err)
				continue
			}
			_, err = fs.eventReceiver.Send(subEventList)
			if err == nil {
				logger.Logger.Error("Event subscribed to mediaserver", subEventList)
				break
			}
			logger.Logger.Error("Initialize MediaServer Failed ", err)
		}

		logger.Logger.Info("FreeSWITCH Connected, subscribed event list ", subEventList)
		go fs.run(cb)
	}()
	return nil
}

func (fs *MsFreeSWITCHGiz) handleFreeSWITCHEvent(ev *eventsocket.Event, cb adapters.MediaServerCallbacker) {
	var err error
	evName := ev.Get("Event-Name")
	callSid := ev.Get("Variable_call_sid")
	if callSid == "" {
		callSid = ev.Get("Unique-ID")
	}
	parentCallSid := ev.Get("Variable_parent_call_sid")
	logger.Logger.Debug("UniqueuID : ", callSid, " Event Name : ", evName)
	evHeader, _ := json.Marshal(ev.Header)
	switch evName {
	case "CHANNEL_PARK":
		err = cb.CallBackPark(callSid, evHeader)
	case "CHANNEL_HANGUP_COMPLETE":
		err = cb.CallBackHangupComplete(callSid, evHeader)
	case "CHANNEL_ANSWER":
		err = cb.CallBackAnswered(callSid, evHeader)
	case "CHANNEL_PROGRESS_MEDIA":
		err = cb.CallBackProgressMedia(callSid, evHeader)
	case "SESSION_HEARTBEAT":
		err = cb.CallBackSessionHeartBeat(parentCallSid, callSid)
	case "RECORD_START":
		err = cb.CallBackRecordingStart(callSid, evHeader)
	case "RECORD_STOP":
		err = cb.CallBackRecordingStop(callSid, evHeader)
	case "CHANNEL_ORIGINATE":
		err = cb.CallBackOriginate(callSid, evHeader)
	case "DTMF":
		err = cb.CallBackDTMFDetected(callSid, evHeader)
	default:
		logger.Logger.Debug("UniqueID : ", callSid, " Event Name : ", evName)
		if err != nil {
			logger.Logger.Error("Error while processing callback events", err)
		}
	}
}

func (fs *MsFreeSWITCHGiz) run(cb adapters.MediaServerCallbacker) {
	for {
		ev, err := fs.eventReceiver.ReadEvent()
		if err != nil {
			logger.Logger.Error("FreeSWITCH ESL Read event failed - Need to debug here", err)
			go func() {
				logger.Logger.Info("Reconnect to the FreeSWITCH again ")
				err = fs.InitializeCallbackMediaServers(cb)
				if err != nil {
					logger.Logger.Error("Error while processing callback events", err)
				}
			}()
			break
		}
		go fs.handleFreeSWITCHEvent(ev, cb)
	}
}

func (fs MsFreeSWITCHGiz) EnableSessionHeartBeat(uuid, interval string) error {
	logger.Logger.Debug("EnableSessionHeartBeat : ", uuid, "remote addres")
	type MSG map[string]string
	ev, err := fs.fsConn.SendMsg(eventsocket.MSG{
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
	_, err := fs.fsConn.SendMsg(eventsocket.MSG{
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
	time.Sleep(1 * time.Second)
	return nil
}

func (fs MsFreeSWITCHGiz) AnswerCall(uuid string) error {
	logger.Logger.Debug("Answering the call with : ", uuid)
	type MSG map[string]string
	_, err := fs.fsConn.SendMsg(eventsocket.MSG{
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
	time.Sleep(1 * time.Second)
	return nil
}

func (fs MsFreeSWITCHGiz) PlayMediaFile(uuid string, fileUrl string, loopCount string) error {
	logger.Logger.Debug("PlayMediaFile : ", uuid, " FileURL: ", fileUrl)
	type MSG map[string]string
	_, err := fs.fsConn.SendMsg(eventsocket.MSG{
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
	_, err := fs.fsConn.SendMsg(eventsocket.MSG{
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
	_, err := fs.fsConn.SendMsg(eventsocket.MSG{
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
	_, err := fs.fsConn.Send(cmd)
	if err != nil {
		logger.Logger.Error("Error while executing the new outbound command", err)
		return err
	}
	return nil
}

func (fs MsFreeSWITCHGiz) CallHangup(uuid string) error {
	logger.Logger.Debug("Hangup : ", uuid)
	_, err := fs.fsConn.SendMsg(eventsocket.MSG{
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
	err := fs.PlayMediaFile(uuid, "silence_stream://2000", "1")
	if err != nil {
		//silence play failed
	}
	logger.Logger.Debug("Hangup : ", uuid)
	ev, err := fs.fsConn.SendMsg(eventsocket.MSG{
		"call-command":     "execute",
		"execute-app-name": "hangup",
		"execute-app-arg":  reason,
		"event-lock":       "true"}, uuid, "")
	logger.Logger.Debug("Destination Event : ", ev, " Error:", err)
	return nil
}

func (fs MsFreeSWITCHGiz) CallHangupWithReason(uuid string, reason string) error {
	err := fs.PlayMediaFile(uuid, "silence_stream://2000", "1")
	if err != nil {
		//silence play failed
	}
	logger.Logger.Debug("Hangup : ", uuid)
	ev, err := fs.fsConn.SendMsg(eventsocket.MSG{
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
	ev, err := fs.fsConn.SendMsg(eventsocket.MSG{
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
	ev, err := fs.fsConn.Send(fmt.Sprintf("bgapi uuid_recv_dtmf %s %s", uuid))
	if err != nil {
		return err
	}
	logger.Logger.Debug("Destination Event : ", ev, " Error:", err)
	return nil
}

func (fs MsFreeSWITCHGiz) BreakAllUuid(uuid string) error {
	logger.Logger.Debug("Break All Command to Uuid  : ", uuid)
	ev, err := fs.fsConn.Send(fmt.Sprintf("bgapi uuid_break %s all", uuid))
	if err != nil {
		return err
	}
	logger.Logger.Debug("Destination Event : ", ev, " Error:", err)
	return nil
}

func (fs MsFreeSWITCHGiz) UuidQueueCount(uuid string) (bool, error) {
	ev, err := fs.fsConn.Send(fmt.Sprintf("api uuid_queue_count %s", uuid))
	if err != nil {
		return false, err
	}
	if ev != nil && strings.HasPrefix(ev.Body, "+OK 0") {
		return true, nil
	}
	return false, nil
}

func (fs MsFreeSWITCHGiz) CallBridge(uuid string, bridgeArgs string) error {
	logger.Logger.Debug("Bridge : ", uuid, " And ", bridgeArgs)
	ev, err := fs.fsConn.SendMsg(eventsocket.MSG{
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
	ev, err := fs.fsConn.SendMsg(eventsocket.MSG{
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
	ev, err := fs.fsConn.SendMsg(eventsocket.MSG{
		"call-command":     "execute",
		"execute-app-name": "set",
		"execute-app-arg":  "RECORD_STEREO=true",
		"event-lock":       "true"}, uuid, "")
	logger.Logger.Debug("Destination Event : ", ev, " Error:", err)
	return err
}

func (fs MsFreeSWITCHGiz) Set(uuid, value string) error {
	logger.Logger.Debug("Set Variable : ", uuid)
	ev, err := fs.fsConn.SendMsg(eventsocket.MSG{
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
	ev, err := fs.fsConn.SendMsg(eventsocket.MSG{
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
	ev, err := fs.fsConn.SendMsg(eventsocket.MSG{
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

func (fs MsFreeSWITCHGiz) Record(uuid string, recordFile string, maxDuration string, silenceSeconds string) error {
	recordString := fmt.Sprintf("%s %s 40 %s", recordFile, maxDuration, silenceSeconds)
	if silenceSeconds == "0" {
		recordString = fmt.Sprintf("'%s %s'", recordFile, maxDuration)
	}
	ev, err := fs.fsConn.SendMsg(eventsocket.MSG{
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
	ev, err := fs.fsConn.SendMsg(eventsocket.MSG{
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
	ev, err := fs.fsConn.SendMsg(eventsocket.MSG{
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
	ev, err := fs.fsConn.SendMsg(eventsocket.MSG{
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

/*
	connection pool for freeswitch but we need to make sure one call will use same connection

*/
