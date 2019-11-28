package my

import (
	"net/http"

	"os"
	"encoding/json"
	"loadgen/lib"
	"io/ioutil"

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
	client:=&http.Client{Timeout: timeoutNs}
	resp,err:=client.Get("http://"+os.Getenv("THOST")+"/latest")
	if err != nil {
		return nil,err
    }
	defer resp.Body.Close()
	x,err:= ioutil.ReadAll(resp.Body)
	return x,err
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
		d.Code=lib.CALL_STATUS_FILED
		d.Msg=msg
	}else{
		d.Code=lib.CALL_STATUS_SUCCESS
	}

	return d

}
