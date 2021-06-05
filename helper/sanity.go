package helper

import (
	"regexp"
	"strconv"
)

func IsValidRecordingStatusCallbackEvent(event string) bool {
	switch event {
	case
		"in-progress",
		"completed",
		"absent":
		return true
	}
	return false
}
func IsValidRecordValue(record string) bool {
	switch record {
	case
		"true",
		"false",
		"record-from-answer",
		"record-from-ringing",
		"record-from-answer-dual",
		"do-not-record",
		"record-from-ringing-dual":
		return true
	}
	return false
}
func IsValidRecordingTrack(track string) bool {
	switch track {
	case
		"inbound",
		"outbound",
		"both":
		return true
	}
	return false
}

func IsValidTrim(trim string) bool {
	switch trim {
	case
		"trim-silence",
		"do-not-trim":
		return true
	}
	return false
}

func IsValidCallReason(reason string) bool {
	return len(reason) <= 50
}

func IsValidTimeOut(timeout string) bool {
	intTimeout, err := strconv.Atoi(timeout)
	if err != nil {
		return false
	}
	if intTimeout > 600 {
		intTimeout = 600
	}
	return true
}

func IsValidTimeLimit(timeLimit string) bool {
	_, err := strconv.Atoi(timeLimit)
	if err != nil {
		return false
	}
	return true
}

func IsValidRingTone(ringtone string) bool {
	switch ringtone {
	case
		"at",
		"au", "bg", "br", "be", "ch",
		"cl", "cn", "cz", "de", "dk",
		"ee", "es", "fi", "fr", "gr",
		"hu", "il", "in", "it", "lt",
		"jp", "mx", "my", "nl", "no",
		"nz", "ph", "pl", "pt", "ru",
		"se", "sg", "th", "uk", "us", "us-old", "tw", "ve", "za":
		return true
	}
	return false
}

func DtmfSanity(number string) string {
	reg, err := regexp.Compile("w#*[^0-9]+")
	if err != nil {
		return number
	}
	return reg.ReplaceAllString(number, "")
}