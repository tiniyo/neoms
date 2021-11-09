package constant

import "errors"

var (
	ErrEmptyXML = errors.New("XML Repsonse is empty")
	ErrInvalidXML = errors.New("XML Response is invalid")
	ErrEmptyRedirect = errors.New("XML Redirect is empty")
	ErrGatherTimeout = errors.New("gather Timeout")
)
