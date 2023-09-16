package main

import (
	"fmt"
	"time"
)

func main() {
	var newTime time.Time
	timeServer := NewVirtualTimeFromDate(newTime)
	timeServer.CabPublishEvents = true
	go timeServer.Bang(timeServer.Now(), timeServer.Interval, timeServer.End)

	// sleep for example
	time.Sleep(1 * time.Second)

	newEvent, err := timeServer.NewEvent()
	if err != nil {
		panic(err)
	}
	timeServer.RegisterEvent(newEvent.ID, newEvent)
	exampleTask(newEvent)
	timeServer.PopEvent(newEvent.ID, newEvent.ID)

	// block
	time.Sleep(5 * time.Second)

}

func exampleTask(event *Event) {
	defer event.CompleteEvent()
	// do something with the event
	time.Sleep(1 * time.Second)
	fmt.Println("event done!")

}
