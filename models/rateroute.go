package models

type OriginationRate struct {
	Rate         float32 `json:"rate_per_minute"`
	InitialPulse int64   `json:"initial_pulse"`
	SubPulse     int64   `json:"sub_pulse"`
}

type TerminationRate struct {
	Rate             float32 `json:"rate_per_minute"`
	InitialPulse     int64   `json:"initial_pulse"`
	SubPulse         int64   `json:"sub_pulse"`
	PrimaryIP        string  `json:"primary_ip"`
	FailoverIP       string  `json:"secondary_ip"`
	Prefix           string  `json:"match_prefix"`
	Priority         int64   `json:"priority"`
	TrunkPrefix      string  `json:"trunk_prefix"`
	RemovePrefix     string  `json:"remove_prefix"`
	SipPilotNumber   string  `json:"sip_pilot_number"`
	FromRemovePrefix string  `json:"from_remove_prefix"`
	Username         string  `json:"username"`
	Password         string  `json:"password"`
}

type InternalRateRoute struct {
	PulseRate            float64
	Pulse                int64
	RoutingGatewayString string
	RoutingUserAuthToken string
	SipPilotNumber       string
	TrunkPrefix          string
	RemovePrefix         string /**/
	FromRemovePrefix     string
}

type RatingRoutingResponse struct {
	Orig OriginationRate    `json:"origination"`
	Term []*TerminationRate `json:"termination"`
}
