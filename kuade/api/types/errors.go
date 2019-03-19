package types

import (
	"encoding/xml"
)

type APIError struct {
	Code           string
	Description    string
	HTTPStatusCode int
}

// APIErrorResponse - error response format
type APIErrorResponse struct {
	XMLName   xml.Name `xml:"Error" json:"-"`
	Code      string
	Message   string
	Resource  string
	RequestID string `xml:"RequestId" json:"RequestId"`
	HostID    string `xml:"HostId" json:"HostId"`
}

// APIErrorCode type of error status.
type APIErrorCode int

const (
	ErrNone APIErrorCode = iota

	ErrAccessDenied
	ErrBadRequest
)
