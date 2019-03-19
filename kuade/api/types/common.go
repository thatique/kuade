package types

import (
	"net/url"
	"strconv"
)

type PaginationArgs struct {
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

func (p PaginationArgs) IsEmpty() bool {
	return p.Limit == 0 && p.Offset == 0
}

func ExtractPaginationArgs(values *url.Values) (pagination PaginationArgs, code APIErrorCode) {
	var (
		err    error
		offset int
		limit  int
	)

	code = ErrNone

	if values.Get("limit") != "" {
		if limit, err = strconv.Atoi(values.Get("limit")); err != nil {
			code = ErrBadRequest
			return
		}
	}

	if values.Get("offset") != "" {
		if offset, err = strconv.Atoi(values.Get("offset")); err != nil {
			code = ErrBadRequest
			return
		}
	}

	pagination = PaginationArgs{Limit: limit, Offset: offset}
	return
}
