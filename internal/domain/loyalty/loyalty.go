package loyalty

import "errors"

type Status string

const (
	REGISTERED Status = "REGISTERED"
	INVALID    Status = "INVALID"
	PROCESSING Status = "PROCESSING"
	PROCESSED  Status = "PROCESSED"
)

var (
	CheckLaterStatuses = []Status{REGISTERED, PROCESSING}
)

var (
	ErrUnavailable         = errors.New("service unavailable")
	ErrWrongResponseStatus = errors.New("wrong response status code")
	ErrWrongContentType    = errors.New("wrong response content type")
)

type AccrualResponse struct {
	Order   string `json:"order"`
	Status  Status `json:"status"`
	Accrual int    `json:"accrual"`
}
