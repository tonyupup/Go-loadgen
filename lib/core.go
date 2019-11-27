package lib

import (
	"time"
)



type Generator interface {
	Start() bool
	Stop() bool
	Status() uint32
	CallCount() uint32
}

type Caller interface {
	BuildReq() RawReq
	Call(req []byte, timeoutNs time.Duration) ([]byte, error)
	CheckResp(rawReq RawReq, rawResp RawResp) *CallResult
}

//GoTickler tickels
type GoTickler interface {
	Take()         //get a tickle
	Return()       //return a tickle
	Active() bool  //is active
	Total() uint32 //total number of tickle
	Remainder() uint32
}