package lib

import (
	"time"
)

const (
	//CALL_STATUS_SUCCESS
	CALL_STATUS_SUCCESS = 0
	CALL_STATUS_TIMEOUT = 1
	CALL_STATUS_FILED   = 2
)

const (
	RET_FAILED=0
	RET_SUCCESS=1
	RET_TIMEOUT=2
)
//RawReq rawreq
type RawReq struct {
	ID   uint32
	Body []byte
}

//RawResp rawresp
type RawResp struct {
	ID     uint32
	Body   []byte
	Err    error
	Elapse time.Duration
}

//CallResult callresult
type CallResult struct {
	ID     uint32
	Req    *RawReq
	Resq   *RawResp
	Code   uint32
	Msg    string
	Elapse time.Duration
}
