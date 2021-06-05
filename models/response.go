package models


type CallResponse struct {
	DateUpdated     interface{} `json:"date_updated,omitempty"`
	PriceUnit       string      `json:"price_unit,omitempty"`
	ParentCallSid   interface{} `json:"parent_call_sid,omitempty"`
	CallerName      interface{} `json:"caller_name,omitempty"`
	Duration        interface{} `json:"duration,omitempty"`
	From            string      `json:"from,omitempty"`
	To              string      `json:"to,omitempty"`
	Annotation      interface{} `json:"annotation,omitempty"`
	AnsweredBy      interface{} `json:"answered_by,omitempty"`
	Sid             string      `json:"sid,omitempty"`
	QueueTime       string      `json:"queue_time,omitempty"`
	Price           interface{} `json:"price,omitempty"`
	APIVersion      string      `json:"api_version,omitempty"`
	Status          string      `json:"status,omitempty"`
	Direction       string      `json:"direction,omitempty"`
	StartTime       interface{} `json:"start_time,omitempty"`
	DateCreated     interface{} `json:"date_created,omitempty"`
	FromFormatted   string      `json:"from_formatted,omitempty"`
	GroupSid        interface{} `json:"group_sid,omitempty"`
	TrunkSid        interface{} `json:"trunk_sid,omitempty"`
	ForwardedFrom   interface{} `json:"forwarded_from,omitempty"`
	URI             string      `json:"uri,omitempty"`
	AccountSid      string      `json:"account_sid,omitempty"`
	EndTime         interface{} `json:"end_time,omitempty"`
	ToFormatted     string      `json:"to_formatted,omitempty"`
	PhoneNumberSid  string      `json:"phone_number_sid,omitempty"`
	SubresourceUris struct {
		Notifications     string `json:"notifications,omitempty"`
		Recordings        string `json:"recordings,omitempty"`
		Payments          string `json:"payments,omitempty"`
		Feedback          string `json:"feedback,omitempty"`
		Events            string `json:"events,omitempty"`
		FeedbackSummaries string `json:"feedback_summaries,omitempty"`
	} `json:"subresource_uris,omitempty"`
}

