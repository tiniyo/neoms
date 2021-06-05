package models

import "time"

type PhoneNumberInfo struct {
	PhoneNumber  string      `json:"phone_number"`
	Rate         float64     `json:"rps"`
	AuthID       string      `json:"acc_id"`
	VendorAuthID string      `json:"parent_act_id"`
	InitPulse    int64       `json:"initial_pulse"`
	SubPulse     int64       `json:"sub_pulse"`
	Application  Application `json:"application"`
	Host         string      `json:"host"`
}

type Application struct {
	Name                 string `json:"AppName"`
	InboundURL           string `json:"Url"`
	InboundMethod        string `json:"Method"`
	FallbackMethod       string `json:"FallbackMethod"`
	FallbackUrl          string `json:"FallbackUrl"`
	StatusCallback       string `json:"StatusCallback"`
	StatusCallbackMethod string `json:"StatusCallbackMethod"`
	StatusCallbackEvent  string `json:"StatusCallbackEvent"`
}
type NumberAPIResponse struct {
	AppResponse PhoneNumberInfo `json:"message"`
}

type SipLocation struct {
	ID           string    `json:"id"`
	Ruid         string    `json:"ruid"`
	Username     string    `json:"username"`
	Domain       string    `json:"domain"`
	Contact      string    `json:"contact"`
	Received     string    `json:"received"`
	Path         string    `json:"path"`
	Expires      time.Time `json:"expires"`
	Q            string    `json:"q"`
	Callid       string    `json:"callid"`
	Cseq         string    `json:"cseq"`
	LastModified time.Time `json:"last_modified"`
	Flags        string    `json:"flags"`
	Cflags       string    `json:"cflags"`
	UserAgent    string    `json:"user_agent"`
	Socket       string    `json:"socket"`
	Methods      string    `json:"methods"`
	Instance     string    `json:"instance"`
	RegID        string    `json:"reg_id"`
	ServerID     string    `json:"server_id"`
	ConnectionID string    `json:"connection_id"`
	Keepalive    string    `json:"keepalive"`
	Partition    string    `json:"partition"`
}
