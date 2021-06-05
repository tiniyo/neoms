package models

type DialAttr struct {
	DialAnswerOnBridge string `json:"DialAnswerOnBridge"`
	DialCallerId       string `json:"DialCallerId"`
	DialHangupOnStar   string `json:"DialHangupOnStar"`
	DialAction         string `json:"DialAction"`
	DialMethod         string `json:"DialMethod"`
	DialRingTone       string `json:"DialRingTone"`
	DialTimeLimit      string `json:"DialTimeLimit"`
	DialTimeout        string `json:"DialTimeout"`
}

type DialRecordAttr struct {
	RecordTimeout            string `json:"RecordTimeout"`
	RecordFinishOnKey        string `json:"RecordFinishOnKey"`
	RecordMaxLength          string `json:"RecordMaxLength"`
	RecordPlayBeep           string `json:"RecordPlayBeep"`
	RecordAction             string `json:"RecordAction"`
	RecordMethod             string `json:"RecordMethod"`
	RecordStorageUrl         string `json:"RecordStorageUrl"`
	RecordStorageUrlMethod   string `json:"RecordStorageUrlMethod"`
	RecordTranscribe         string `json:"RecordTranscribe"`
	RecordTranscribeCallback string `json:"RecordTranscribeCallback"`
}

type DialSipAttr struct {
	DialSipMethod               string `json:"DialSipMethod"`
	DialSipPassword             string `json:"DialSipPassword"`
	DialSipStatusCallbackEvent  string `json:"DialSipStatusCallbackEvent"`
	DialSipStatusCallback       string `json:"DialSipStatusCallback"`
	DialSipStatusCallbackMethod string `json:"DialSipStatusCallbackMethod"`
	DialSipUrl                  string `json:"DialSipUrl"`
	DialSipUsername             string `json:"DialSipUsername"`
}

type DialNumberAttr struct {
	DialNumberMethod               string `json:"DialNumberMethod"`
	DialNumberSendDigits           string `json:"DialNumberSendDigits"`
	DialNumberStatusCallbackEvent  string `json:"DialNumberStatusCallbackEvent"`
	DialNumberStatusCallback       string `json:"DialNumberStatusCallback"`
	DialNumberStatusCallbackMethod string `json:"DialNumberStatusCallbackMethod"`
	DialNumberUrl                  string `json:"DialNumberUrl"`
	DialNumberByoc                 string `json:"DialNumberByoc"`
}

type GatherAttr struct {
	GatherFinishOnKey         string `json:"GatherFinishOnKey"`
	GatherTimeout             string `json:"GatherTimeout"`
	GatherAction              string `json:"GatherAction"`
	GatherMethod              string `json:"GatherMethod"`
	GatherNumDigit            string `json:"GatherNumDigit"`
	GatherActionOnEmptyResult string `json:"GatherActionOnEmptyResult"`
	GatherEnhanced            string `json:"GatherEnhanced"`
}
