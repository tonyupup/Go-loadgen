package my

import (
	"os"
	"encoding/json"
	"loadgen/lib"
	"io/ioutil"
	"net"
	"sync/atomic"
	"time"
)

type myCaller struct {
	id uint32
}

//NewCaller reuturn a new Caller
func NewCaller() (lib.Caller, error) {
	ca := &myCaller{}
	return ca, nil
}
func (c *myCaller) BuildReq() lib.RawReq {
	d := lib.RawReq{ID: c.id}
	atomic.AddUint32(&c.id, 1)
	return d
}

func (c *myCaller) Call(req []byte, timeoutNs time.Duration) ([]byte, error) {
	client, err := net.DialTimeout("tcp", os.Getenv("THOST"), timeoutNs)
	if err != nil {
		return nil, err
	}
	client.Write([]byte("GET /latest HTTP/1.1\n"))
	defer client.Close()
	return ioutil.ReadAll(client)
}

func (c *myCaller) CheckResp(rawReq lib.RawReq, rawResp lib.RawResp) *lib.CallResult {
	var m interface{}
	var msg string
	err := json.Unmarshal(rawResp.Body, &m)
	d := &lib.CallResult{
			ID:     rawReq.ID,
			Req:    &rawReq,
			Resq:   &rawResp,
			Elapse: rawResp.Elapse,
		}
	if err != nil {
		msg = err.Error()
	}
	if msg!=""{
		d.Code=lib.CALL_STATUS_FILED
		d.Msg=msg
	}else{
		d.Code=lib.CALL_STATUS_SUCCESS
	}

	return d

}
