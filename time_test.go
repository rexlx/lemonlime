package main

import (
	"testing"
	"time"
)

func Test_NewVirtualTimeFromNow(t *testing.T) {
	timeServer := NewVirtualTimeFromNow()

	// check that the time is now
	if timeServer.Elapsed != 0 {
		t.Errorf("Expected %v, got %v", 0, timeServer.Elapsed)
	}
}

func Test_NewVirtualTimeFromDate(t *testing.T) {
	timeServer := NewVirtualTimeFromDate(time.Date(2007, 8, 01, 01, 01, 01, 00, time.Local))
	timeServer.CabPublishEvents = true
	go timeServer.Bang(timeServer.Now(), 86400*time.Second, time.Date(2010, 8, 01, 01, 01, 01, 00, time.Local))
	time.Sleep(2 * time.Second)
	// check that the time is now
	if timeServer.Now() != time.Date(2007, 10, 18, 01, 01, 01, 00, time.Local) {
		t.Errorf("Expected %v, got %v", time.Date(2007, 10, 18, 01, 01, 01, 00, time.Local), timeServer.Now())
	}
	if timeServer.Elapsed != 1872*time.Hour {
		t.Errorf("Expected %v, got %v -> %v", 0, timeServer.Elapsed, timeServer.Now())
	}
}
