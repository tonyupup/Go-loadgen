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
	c:=make(chan *lib.CallResult,500000)
	x,err:=my.NewGenerator(call,time.Second*5,100,time.Second*10,c)
	if err!=nil{
		fmt.Println(err)
		return
	}
	x.Start()

	for c :=range c{
		fmt.Println(c.Msg)
	}
}
