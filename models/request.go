package models

/*
   authid -> call_uuid
	 call_uuid -> call_request
	 get auth id from the response and from that response.
*/

type CallRequest struct {
	AsyncAmdStatusCallbackMethod       string  `json:"AsyncAmdStatusCallbackMethod"`
	AsyncAmdStatusCallback             string  `json:"AsyncAmdStatusCallback"`
	AsyncAmd                           string  `json:"AsyncAmd"`
	MachineDetectionSilenceTimeout     string  `json:"MachineDetectionSilenceTimeout"`
	MachineDetectionSpeechEndThreshold string  `json:"MachineDetectionSpeechEndThreshold"`
	MachineDetectionSpeechThreshold    string  `json:"MachineDetectionSpeechThreshold"`
	MachineDetectionTimeout            string  `json:"MachineDetectionTimeout"`
	MachineDetection                   string  `json:"MachineDetection"`
	SipAuthPassword                    string  `json:"SipAuthPassword"`
	SipAuthUsername                    string  `json:"SipAuthUsername"`
	LoopPlay                           string  `json:"loop_play" example:"3"`
	Timeout                            string  `json:"Timeout"`
	From                               string  `json:"From" example:"15677654321"`
	To                                 string  `json:"To" example:"15677654321"`
	CallerName                         string  `json:"caller_name" example:"Tiniyo"`
	CallerId                           string  `json:"CallerId"`
	Byoc                               string  `json:"Byoc"`
	CallReason                         string  `json:"CallReason"`
	Trim                               string  `json:"Trim"`
	RecordingStatusCallbackEvent       string  `json:"RecordingStatusCallbackEvent"`
	RecordingTrack                     string  `json:"RecordingTrack"`
	RecordingChannels                  string  `json:"RecordingChannels"`
	ParentCallSid                      string  `json:"parent_call_sid"`
	AccountSid                         string  `json:"AccountSid"`
	Record                             string  `json:"Record"`
	SendDigits                         string  `json:"SendDigits"`
	Play                               string  `json:"play" example:"https://tiniyo.s3.amazonaws.com/MissionImpossible.mp3"`
	Speak                              string  `json:"speak" example:"Hello Dear, Thanks for using our service"`
	ApplicationSid                     string  `json:"ApplicationSid" example:"your tiniyo application id"`
	TinyML                             string  `json:"TinyML" example:"<Response><Say>Hello World</Say>"`
	Url                                string  `json:"Url" example:"https://raw.githubusercontent.com/tiniyo/public/master/answer.xml"`
	Method                             string  `json:"Method" example:"GET"`
	FallbackMethod                     string  `json:"FallbackMethod"`
	FallbackUrl                        string  `json:"FallbackUrl"`
	StatusCallback                     string  `json:"StatusCallback"`
	StatusCallbackMethod               string  `json:"StatusCallbackMethod"`
	StatusCallbackEvent                string  `json:"StatusCallbackEvent"`
	RecordingStatusCallback            string  `json:"RecordingStatusCallback"`
	RecordingStatusCallbackMethod      string  `json:"RecordingStatusCallbackMethod"`
	Rate                               float64 `json:"rate"`
	Pulse                              int64   `json:"pulse"`
	MaxDuration                        int64   `json:"max_duration"`
	DestType                           string  `json:"DestType"`
	SrcType                            string  `json:"SrcType"`
	VendorAuthID                       string  `json:"ParentAuthId"`
	SipPilotNumber                     string  `json:"SipPilotNumber"`
	Sid                                string  `json:"Sid"`
	Bridge                             string  `json:"Bridge"`
	Host                               string  `json:"Host"`
	SrcDirection                       string  `json:"SrcDirection"`
	IsCallerId                         string  `json:"IsCallerId"`
	SipTrunk                           string  `json:"SipTrunk"`
	FromRemovePrefix				   string  `json:"FromRemovePrefix"`
	DialAttr
	DialRecordAttr
	Callback
	RecordCallback
	CallResponse
	DialSipAttr
	DialNumberAttr
	GatherAttr
}

type CallUpdateRequest struct {
	Sid                  string `json:"Sid,omitempty"`
	Url                  string `json:"Url"`
	Method               string `json:"Method" example:"GET"`
	FallbackMethod       string `json:"FallbackMethod"`       // `json:"FallbackMethod"`
	FallbackUrl          string `json:"FallbackUrl"`          // `json:"FallbackUrl"`
	StatusCallback       string `json:"StatusCallback"`       // `json:"StatusCallback"`
	StatusCallbackMethod string `json:"StatusCallbackMethod"` // `json:"StatusCallbackMethod"`
	Status               string `json:"Status"`               // `json:"StatusCallback"`
}
type CallPlayRequest struct {
	Urls   string `json:"urls"`
	Length int    `json:"length"`
	Legs   string `json:"legs"`
	Loop   int    `json:"loop"`
	Mix    bool   `json:"mix"`
}

type CallSpeakRequest struct {
	Text     string `json:"text"`
	Voice    string `json:"voice"`
	Language string `json:"language"`
	Legs     string `json:"legs"`
	Loop     bool   `json:"loop"`
	Mix      bool   `json:"mix"`
}

type CallEventRequest struct {
	SendDigits     string `json:"digits"`
	DigitsReceived string `json:"digitsReceived"`
	Leg            string `json:"leg"`
}

type Call struct {
	uuid string
	cr   CallRequest
}
