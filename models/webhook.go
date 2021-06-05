package models

type (
	Callback struct {
		Called           string `json:"Called,omitempty" form:"Called" query:"Called" freeswitch:"Caller-Destination-Number"`
		Direction        string `json:"Direction,omitempty" form:"Direction" query:"Direction" freeswitch:"Variable_direction"`
		Timestamp        string `json:"Timestamp,omitempty" form:"Timestamp" query:"Timestamp" freeswitch:"Event-Date-GMT"`
		CallSid          string `json:"CallSid,omitempty" form:"CallSid" query:"CallSid" freeswitch:"Variable_call_sid"`
		To               string `json:"To,omitempty" form:"To" query:"To" freeswitch:"Variable_sip_req_user"`
		AccountSid       string `json:"AccountSid,omitempty" form:"AccountSid" query:"AccountSid" freeswitch:"Variable_tiniyo_accid"`
		Caller           string `json:"Caller,omitempty" form:"Caller" query:"Caller" freeswitch:"Caller-ANI"`
		From             string `json:"From,omitempty" form:"From" query:"From" freeswitch:"Variable_sip_from_user"`
		ParentCallSid    string `json:"ParentCallSid,omitempty" form:"ParentCallSid" query:"ParentCallSid" freeswitch:"Variable_parent_call_sid"`
		CallStatus       string `json:"CallStatus,omitempty" form:"CallStatus" query:"CallStatus" freeswitch:"Hangup-Cause"`
		CallDuration     string `json:"CallDuration,omitempty" form:"CallDuration" query:"CallDuration" freeswitch:"Variable_billsec"`
		ToState          string `json:"ToState,omitempty" form:"ToState" query:"ToState"`
		CallerCountry    string `json:"CallerCountry,omitempty" form:"CallerCountry" query:"CallerCountry" `
		CallbackSource   string `json:"CallbackSource,omitempty" form:"CallbackSource" query:"CallbackSource"`
		SipResponseCode  string `json:"SipResponseCode,omitempty" form:"SipResponseCode" query:"SipResponseCode"`
		CallerState      string `json:"CallerState,omitempty" form:"CallerState" query:"CallerState"`
		ToZip            string `json:"ToZip,omitempty" form:"ToZip" query:"ToZip"`
		SequenceNumber   string `json:"SequenceNumber,omitempty" form:"SequenceNumber" query:"SequenceNumber"`
		CallerZip        string `json:"CallerZip,omitempty" form:"CallerZip" query:"CallerZip"`
		ToCountry        string `json:"ToCountry,omitempty" form:"ToCountry" query:"ToCountry"`
		CalledZip        string `json:"CalledZip,omitempty" form:"CalledZip" query:"CalledZip"`
		ApiVersion       string `json:"ApiVersion,omitempty" form:"ApiVersion" query:"ApiVersion"`
		CalledCity       string `json:"CalledCity,omitempty" form:"CalledCity" query:"CalledCity"`
		CalledCountry    string `json:"CalledCountry,omitempty" form:"CalledCountry" query:"CalledCountry"`
		CallerCity       string `json:"CallerCity,omitempty" form:"CallerCity" query:"CallerCity"`
		ToCity           string `json:"ToCity,omitempty" form:"ToCity" query:"ToCity"`
		FromCountry      string `json:"FromCountry,omitempty" form:"FromCountry" query:"FromCountry"`
		FromCity         string `json:"FromCity,omitempty" form:"FromCity" query:"FromCity"`
		CalledState      string `json:"CalledState,omitempty" form:"CalledState" query:"CalledState"`
		FromZip          string `json:"FromZip,omitempty" form:"FromZip" query:"FromZip"`
		FromState        string `json:"FromState,omitempty" form:"FromState" query:"FromState"`
		InitiationTime   string `json:"InitiationTime,omitempty" form:"InitiationTime" query:"InitiationTime"`
		AnswerTime       string `json:"AnswerTime,omitempty" form:"AnswerTime" query:"AnswerTime"`
		RingTime         string `json:"RingTime,omitempty" form:"RingTime" query:"RingTime"`
		HangupTime       string `json:"HangupTime,omitempty" form:"HangupTime" query:"HangupTime"`
		Digits           string `json:"Digits,omitempty" form:"Digits" query:"Digits"`
		DtmfInputType    string `json:"DtmfInputType,omitempty" form:"DtmfInputType" query:"DtmfInputType"`
		DialCallStatus   string `json:"DialCallStatus,omitempty"`
		DialCallSid      string `json:"DialCallSid,omitempty"`
		DialCallDuration string `json:"DialCallDuration,omitempty"`
		RecordingUrl     string `json:"RecordingUrl,omitempty" form:"RecordingUrl" query:"RecordingUrl" freeswitch:"Variable_tiniyo_recording_file"`
		PriceUnit        string `json:"price_unit,omitempty"`
	}


	RecordCallback struct {
		RecordEventTimestamp string `json:"Timestamp,omitempty" form:"Timestamp" query:"Timestamp" freeswitch:"Event-Date-GMT"`
		RecordingSource      string `json:"RecordingSource,omitempty" form:"RecordingSource" query:"RecordingSource"`
		RecordingTrack       string `json:"RecordingTrack,omitempty" form:"RecordingTrack" query:"RecordingTrack"`
		RecordingSid         string `json:"RecordingSid,omitempty" form:"RecordingSid" query:"RecordingSid"`
		RecordingUrl         string `json:"RecordingUrl,omitempty" form:"RecordingUrl" query:"RecordingUrl" freeswitch:"Record-File-Path"`
		RecordingStatus      string `json:"RecordingStatus,omitempty" form:"RecordingStatus" query:"RecordingStatus"`
		RecordingChannels    string `json:"RecordingChannels,omitempty" form:"RecordingChannels" query:"RecordingChannels"`
		ErrorCode            string `json:"ErrorCode,omitempty" form:"ErrorCode" query:"ErrorCode"`
		RecordCallSid        string `json:"CallSid,omitempty" form:"CallSid" query:"CallSid"`
		RecordingStartTime   string `json:"RecordingStartTime,omitempty" form:"RecordingStartTime" query:"RecordingStartTime"`
		RecordAccountSid     string `json:"AccountSid,omitempty" form:"AccountSid" query:"AccountSid"`
		RecordingDuration    string `json:"RecordingDuration,omitempty" form:"RecordingDuration" query:"RecordingDuration" freeswitch:"Variable_record_seconds"`
	}

	DialActionUrlCallback struct {
	}
)
