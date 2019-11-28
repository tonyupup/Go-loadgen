package main

import (
	"loadgen/my"
	"loadgen/lib"
	"fmt"
	"time"
)

func main() {
	call,err:=my.NewCaller()
	if err!=nil{
		fmt.Println(err)
		return 
	}
	c:=make(chan *lib.CallResult,500)
	x,err:=my.NewGenerator(call,time.Second*5,1000,time.Second*10,c)
	if err!=nil{
		fmt.Println(err)
		return
	}
	x.Start()
	count:=make(map[uint32]int)
	for f :=range c{
		count[f.Code]=count[f.Code]+1
	}
	secccall:=count[lib.CALL_STATUS_SUCCESS]
	tps:=float64(secccall)/float64(time.Second*10/1e9)
	fmt.Printf("lps %v,TPS:%0.2f",time.Second*5,tps)
}
