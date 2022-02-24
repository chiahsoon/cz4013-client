package api

import (
	"sync/atomic"
	"time"
)

var RSN *int64 = new(int64)

type Request struct {
	RSN    int
	Method string
	Data   interface{}
	SentAt time.Time
}

func NewRequest() Request {
	req := Request{RSN: int(*RSN)}
	atomic.AddInt64(RSN, 1)
	req.SentAt = time.Now() // !REVIEW
	return req
}

func (req *Request) GetRSN() int {
	return int(req.RSN)
}
