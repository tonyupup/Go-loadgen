package lib

import (
	"fmt"
)



type myGoTickle struct {
	total    uint32
	avtive   bool
	ticketCh chan struct{}
}

//init goTickets
func (gt *myGoTickle) init(total uint32) bool {
	if gt.avtive {
		return false
	}
	ch := make(chan struct{}, total)
	for i := 0; i < int(total); i++ {
		ch <- struct{}{}
	}
	gt.total = total
	gt.avtive = true
	gt.ticketCh = ch
	return true
}

//Take table a tickle
func (gt *myGoTickle) Take() {
	<-gt.ticketCh
}

//Return return
func (gt *myGoTickle) Return() {
	gt.ticketCh <- struct{}{}
}

//Active return is active
func (gt *myGoTickle) Active() bool {
	return gt.avtive
}

//Total reutrn total number of gotickkles
func (gt *myGoTickle) Total() uint32 {
	return gt.total
}

//Remainder reuturn remainder for goticlets
func (gt *myGoTickle) Remainder() uint32 {
	return gt.total - uint32(len(gt.ticketCh))
}

//NewGoTickle return new go tickels
func NewGoTickle(total uint32) (GoTickler, error) {
	gt := &myGoTickle{}
	if !gt.init(total) {
		return nil, fmt.Errorf("The gourouting pool can not be initialized! total=%d", total)
	}
	return gt, nil
}
