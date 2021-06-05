package models

import "net/http"

type RequestError struct {
	StatusCode int
	Err error
}

func (r *RequestError) Error() string {
	return r.Err.Error()
}

func (r *RequestError) RatingRoutingMissing() bool {
	return r.StatusCode == http.StatusServiceUnavailable
}

func (r *RequestError) NestedDialElement() bool {
	return r.StatusCode == http.StatusMethodNotAllowed
}

func (r *RequestError) PaymentRequired() bool {
	return r.StatusCode == http.StatusPaymentRequired
}

func (r *RequestError) BadCallerID() bool {
	return r.StatusCode == http.StatusBadRequest
}