package adapters

type MediaServer interface {
	// Initialize
	InitializeCallbackMediaServers(cb MediaServerCallbacker) error
	//Call
	AnswerCall(uuid string) error
	PreAnswerCall(uuid string) error
	PlayMediaFile(uuid string, fileUrl string, loopCount string) error
	PlayBeep(uuid string) error
	Speak(uuid string, voiceId, text string) error
	CallNewOutbound(cmd string) error
	CallTransfer() error
	CallSendDTMF(uuid string, dtmf string) error
	BreakAllUuid(uuid string) error
	CallReceiveDTMF(uuid string) error
	SetRecordStereo(uuid string) error
	Set(uuid, value string) error
	UuidQueueCount(uuid string) (bool,error)
	MultiSet(uuid, value string) error
	CallRecord(uuid string, recordFile string) error
	Record(uuid string, recordFile string, maxDuration string, silenceSeconds string) error
	CallBridge(uuid string, otherUuid string) error
	CallIntercept(uuid string, otherUuid string) error
	CallHangup(uuid string) error
	CallHangupWithReason(uuid string, reason string) error
	CallHangupWithSync(uuid string, reason string) error
	EnableSessionHeartBeat(uuid, interval string) error
	//Conference
	ConfCreate(uuid, conferenceName string) error
	ConfBridge(uuid, bridgeArgs string) error
	ConfSetAutoCall(uuid, bridgeArgs string) error
	ConfAddMember() error
	ConfRemoveMember() error
}

// MediaServerCallbackInterface callback of the media server
type MediaServerCallbacker interface {
	//Status
	CallBackMediaServerStatus(status int) error
	CallBackDTMFDetected(uuid string, evHeader []byte) error
	CallBackProgress(uuid string) error
	CallBackAnswered(uuid string, evHeader []byte) error
	CallBackProgressMedia(uuid string, evHeader []byte) error
	CallBackHangup(uuid string) error
	CallBackPark(uuid string, evHeader []byte) error
	CallBackDestroy(uuid string) error
	CallBackExecuteComplete(uuid string) error
	CallBackHangupComplete(uuid string, evHeader []byte) error
	CallBackRecordingStart(uuid string,evHeader []byte) error
	CallBackRecordingStop(uuid string,evHeader []byte) error
	CallBackBridged(uuid string) error
	CallBackUnBridged(uuid string) error
	CallBackSessionHeartBeat(puuid, uuid string) error
	CallBackMessage(uuid string) error
	CallBackCustom(uuid string) error
	CallBackOriginate(uuid string,evHeader []byte) error
}
