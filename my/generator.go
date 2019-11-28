package my

import (
	"loadgen/lib"
	"context"
	"fmt"
	"log"
	"os"
	"sync/atomic"
	"time"
)

var loger *log.Logger

func init() {
	loger = log.New(os.Stdout, "INFO ", log.Lshortfile|log.Ltime)
}

const (
	//STATUSORIGINAL 准备阶段
	STATUSORIGINAL uint32 = 0
	//STATUSSTARTING 启动中
	STATUSSTARTING uint32 = 1
	//STATUSSTARTED 运行中
	STATUSSTARTED uint32 = 2
	//STATUSSTOPPING 停止中
	STATUSSTOPPING uint32 = 3
	//STATUSSTOPEND 已停止
	STATUSSTOPEND uint32 = 4
)

//myGenerator
type myGenerator struct {
	caller      lib.Caller           //调用器
	timeoutNs   time.Duration        //处理超时时间 ns
	lps         uint32               //每秒负载均衡
	durationNs  time.Duration        //负载持续时间
	concurrency uint32               //载荷并发量
	tickets     lib.GoTickler        //goroutine 票池
	ctx         context.Context      //上下文
	cancelFunc  context.CancelFunc   //取消函数
	status      uint32               //状态
	callCount   uint32               //调用计数
	resultCh    chan *lib.CallResult // 调用结果
}

//init initlized a myGenerator
func (g *myGenerator) init() error {
	var err error
	g.concurrency = uint32(g.timeoutNs)/(1e9/g.lps) + 1
	g.tickets, err = lib.NewGoTickle(g.concurrency)
	if err != nil {
		return err
	}
	return nil
}

//callOne 调用一个req
func (g *myGenerator) callOne(req *lib.RawReq) *lib.RawResp {
	atomic.AddUint32(&g.callCount, 1)
	if req == nil {
		return &lib.RawResp{ID: 0, Err: fmt.Errorf("Invaid raw reequest")}
	}
	var result lib.RawResp
	start := time.Now().UnixNano()
	Body, err := g.caller.Call(req.Body, g.timeoutNs)
	end := time.Now().UnixNano()
	elapse:=time.Duration(end - start)
	if err != nil {
		result.ID = req.ID
		result.Err = err
		result.Elapse = elapse
	} else {
		result.ID = req.ID
		result.Body = Body
		result.Elapse = elapse
	}
	return &result
}

//sendResult send
func (g *myGenerator) sendResult(result *lib.CallResult) bool {
	if atomic.LoadUint32(&g.status) != STATUSSTARTED {
		//load generator has been stopped.
		return false
	}
	select {
	case g.resultCh <- result:
		return true
	default: //默认选项，防止阻塞
		return false
	}
}

//asyncCall call with async
func (g *myGenerator) asyncCall() {
	g.tickets.Take()
	go func() {
		rawReq := g.caller.BuildReq()
		//callStatus staus 0 calling ,1:call successfull,2:call failed
		var callStatus uint32
		timer := time.AfterFunc(g.timeoutNs, func() {
			//if compare failed ,the callStatus has changed 1
			if !atomic.CompareAndSwapUint32(&callStatus, 0, 2) {
				return
			}
			result := &lib.CallResult{
				ID:     rawReq.ID,
				Req:    &rawReq,
				Code:   lib.CALL_STATUS_TIMEOUT,
				Msg:    fmt.Sprintf("Timeout! (excepted:<%v)!", g.timeoutNs),
				Elapse: g.timeoutNs,
			}
			g.sendResult(result)
		})
		rawResp := g.callOne(&rawReq)
		if !atomic.CompareAndSwapUint32(&callStatus, 0, 1) {
			//比较失败，已经调用超时了，直接Return
			return
		}
		timer.Stop()

		var result *lib.CallResult
		if rawResp.Err != nil {
			result = &lib.CallResult{
				ID:     rawResp.ID,
				Req:    &rawReq,
				Code:  	lib.CALL_STATUS_FILED,
				Msg:    rawResp.Err.Error(),
				Elapse: rawResp.Elapse,
			}
		} else {
			result = g.caller.CheckResp(rawReq, *rawResp)
		}
		g.sendResult(result)
		defer func() {
			if p := recover(); p != nil {
				err, ok := interface{}(p).(error)
				var errMsg string
				if ok {
					errMsg = fmt.Sprintf("Async Cann panic! (error:%s)", err)
				} else {
					errMsg = fmt.Sprintf("Async Call Panic! (clue:%v)", p)
				}
				loger.Println(errMsg)
				result := &lib.CallResult{
					ID:   0,
					Code: lib.CALL_STATUS_FILED,
					Msg:  errMsg,
				}
				g.sendResult(result)
			}
			g.tickets.Return()
		}()
	}()
}

//genLoad 产生载荷并且发送
func (g *myGenerator) genLoad(throttle <-chan time.Time) {
	for {
		//为了及时收到上下文
		select {
		case <-g.ctx.Done():
			g.preStop(g.ctx.Err())
			return
		default:
		}
		//异步调用载荷发送
		g.asyncCall()
		if g.lps > 0 {
			select {
			case <-throttle:
			case <-g.ctx.Done():
				g.preStop(g.ctx.Err())
				return
			}
		}
	}
}

//preStop 预停止
func (g *myGenerator) preStop(ctxError error) {
	loger.Printf("Prepare to stop load generator (cause :%s)\n", ctxError)
	atomic.CompareAndSwapUint32(&g.status, STATUSSTARTED, STATUSSTOPPING)
	close(g.resultCh)
	loger.Println("Close result channel Done.")
	atomic.StoreUint32(&g.status, STATUSSTOPEND)
}
func (g *myGenerator) Start() bool {
	//throttle节流阀
	var throttle <-chan time.Time
	if g.lps > 0 {

		interval := time.Duration(1e9 / g.lps)
		loger.Printf("Setting throttle %v..\n", interval)
		throttle = time.Tick(interval)
	} else {
		return false
	}
	g.ctx, g.cancelFunc = context.WithTimeout(context.Background(), g.durationNs) //设置超时上下文
	g.callCount = 0
	atomic.StoreUint32(&g.status, STATUSSTARTING)
	go func() {
		loger.Println("Generating load...")
		g.genLoad(throttle)
		loger.Printf("Stopped, call count :%d\n", g.callCount)
	}()
	atomic.StoreUint32(&g.status,STATUSSTARTED)
	return true
}

func (g *myGenerator) Stop() bool {
	if !atomic.CompareAndSwapUint32(&g.status, STATUSSTARTED, STATUSSTOPPING) {
		return false
	}
	g.cancelFunc()
	for {
		if atomic.LoadUint32(&g.status) == STATUSSTOPEND {
			break
		}
		time.Sleep(time.Second)
	}
	return true
}
func (g *myGenerator) Status() uint32 {
	return g.status
}
func (g *myGenerator) CallCount() uint32 {
	return g.callCount
}

//NewGenerator return a new generator implete Generator interrface
func NewGenerator(caller lib.Caller, timeoutNs time.Duration, lps uint32, durationNs time.Duration, resultCh chan *lib.CallResult) (lib.Generator, error) {
	gen := &myGenerator{
		caller:     caller,
		timeoutNs:  timeoutNs,
		lps:        lps,
		status:     STATUSORIGINAL,
		durationNs: durationNs,
		resultCh:   resultCh,
	}
	if err := gen.init(); err != nil {
		return nil, err
	}
	return gen, nil
}
