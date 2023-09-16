package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
)

type VirtualTimeInterface interface {
	Now() time.Time
	TotalElapsed() time.Duration
	Advance()
	Bang(start time.Time, interval time.Duration, end time.Time)
}

// Synchronizer is the main struct for the lemonlime package. it implements the VirtualTimeInterface
type Synchronizer struct {
	// logging stream
	Log *log.Logger `json:"-"`

	// Mutex is for locking the members map
	Mu *sync.RWMutex `json:"-"`

	// Time is an implementation of the VirtualTimeInterface
	Time VirtualTimeInterface `json:"-"`

	// Kill is a channel that can be used in a select statement to break a loop
	Kill chan interface{} `json:"-"`

	// subscribers to the synchronizer
	Members map[string][]*Event `json:"members"`

	// the start time of the synchronizer
	Start time.Time `json:"start"`

	// the current time of the synchronizer
	Current time.Time `json:"current"`

	// the total elapsed time of the synchronizer in virtual time
	Elapsed time.Duration `json:"elapsed"`

	// the interval the synchronizer advances time by
	Interval time.Duration `json:"interval"`

	// the end time for the synchronizer
	End time.Time `json:"end"`

	// stops server from accepting new events
	CabPublishEvents bool `json:"wait"`
}

// tests for the presence of any events in any member
func (s *Synchronizer) CanAdvance() bool {
	s.Mu.RLock()
	defer s.Mu.RUnlock()
	for _, events := range s.Members {
		if len(events) != 0 {
			return false
		}
	}
	return true
}

func (s *Synchronizer) ClearEvents(id string) {
	s.Mu.RLock()
	defer s.Mu.RUnlock()
	if _, ok := s.Members[id]; !ok {
		s.Log.Println("ClearEvents: Member not found:", id)
	} else {
		s.Members[id] = []*Event{}
	}
}

// Does NOT maintain order
func (s *Synchronizer) PopEvent(memberId, eventId string) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	if _, ok := s.Members[memberId]; !ok {
		s.Log.Println("PopEvent: Member not found:", memberId)
	} else {
		for i, v := range s.Members[memberId] {
			// TODO maybe we dont care if it's complete?
			if v.ID == eventId && v.Complete {
				s.Log.Println("PopEvent: removing:", eventId)
				s.Members[memberId] = append(s.Members[memberId][:i], s.Members[memberId][i+1:]...)
				return
			}
		}
		s.Log.Println("PopEvent: Event not found:", eventId)
	}
}

// retrieve member's events from the member map
func (s *Synchronizer) GetMember(id string) ([]*Event, error) {
	s.Mu.RLock()
	defer s.Mu.RUnlock()
	if _, ok := s.Members[id]; !ok {
		s.Log.Println("Member not found:", id)
	} else {
		return s.Members[id], nil
	}
	return []*Event{}, fmt.Errorf("member not found: %s", id)
}

// register a member with the synchronizers member map
func (s *Synchronizer) RegisterMember(id string) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	s.Members[id] = make([]*Event, 0)
	fmt.Println("RegisterMember", s.Members)
}

// register an event to a member in the synchronizers member map
// TODO in the event that the member is not found, should we create it?
func (s *Synchronizer) RegisterEvent(id string, event *Event) {
	fmt.Println("RegisterEvent", id, event)
	s.Mu.Lock()
	defer s.Mu.Unlock()
	if _, ok := s.Members[id]; !ok {
		s.Log.Println("Member not found:", id, "registering...")
		s.Members[id] = make([]*Event, 0)
		s.Members[id] = append(s.Members[id], event)
	} else {
		s.Members[id] = append(s.Members[id], event)
	}
}

func (s *Synchronizer) NewEvent() (*Event, error) {
	if !s.CabPublishEvents {
		return nil, fmt.Errorf("cannot create new event: CabPublishEvents == false")
	}
	// create a random uuid with uuid Must
	id := uuid.Must(uuid.NewRandom()).String()
	return &Event{
		ID:          id,
		RealTime:    time.Now(),
		VirtualTime: s.Now(),
	}, nil
}

// just another test function :)
// func (s *Synchronizer) Sleep(in int) {
// 	// var e time.Duration
// 	var t time.Time
// 	start := time.Now()
// 	fmt.Println(t, start)
// 	time.Sleep(time.Duration(in) * time.Second)
// 	newDuration := time.Since(start) * time.Duration(factor)
// 	u := t.Add(newDuration)
// 	fmt.Println(newDuration, u, time.Since(start))
// }

// implement the VirtualTimeInterface
func (s *Synchronizer) Now() time.Time {
	return s.Current
}

// implement the VirtualTimeInterface
func (s *Synchronizer) TotalElapsed() time.Duration {
	return s.Elapsed
}

// implement the VirtualTimeInterface
func (s *Synchronizer) Advance() {
	s.Current = s.Current.Add(s.Interval)
	s.Elapsed = s.Elapsed + s.Interval
}

func (s *Synchronizer) Bang(start time.Time, interval time.Duration, end time.Time) {
	s.Start = start
	s.Current = start
	s.Interval = interval
	s.End = end

eternity:
	for s.Current.Before(s.End) {
		select {
		case <-s.Kill:
			s.Log.Println("kill")
			break eternity
		default:
			// advance the synchronizer if permitted
			if s.CanAdvance() {
				s.Advance()
				s.Log.Println("Time / Elapsed", s.Current, s.Elapsed)
			} else {
				s.Log.Println("Can't advance")
			}
			time.Sleep(time.Millisecond * 25)
		}
	}
}

// NewVirtualTimeFromNow creates a new VirtualTimeInterface from the current time
func NewVirtualTimeFromNow() *Synchronizer {
	members := make(map[string][]*Event)
	ch := make(chan interface{})
	l := log.New(os.Stdout, "syncengine > ", log.Ldate|log.Ltime)
	mu := sync.RWMutex{}
	return &Synchronizer{
		Members:  members,
		Mu:       &mu,
		Kill:     ch,
		Log:      l,
		Start:    time.Now(),
		Current:  time.Now(),
		Elapsed:  0,
		Interval: time.Second,
		End:      time.Now().Add(time.Hour * 24),
	}
}

// NewVirtualTimeFromDate creates a new VirtualTimeInterface from a given time.Time
func NewVirtualTimeFromDate(start time.Time) *Synchronizer {
	members := make(map[string][]*Event)
	ch := make(chan interface{})
	l := log.New(os.Stdout, "syncengine > ", log.Ldate|log.Ltime)
	mu := sync.RWMutex{}
	return &Synchronizer{
		Members:  members,
		Log:      l,
		Mu:       &mu,
		Kill:     ch,
		Start:    start,
		Current:  start,
		Elapsed:  0,
		Interval: time.Second,
		End:      start.Add(time.Hour * 24),
	}
}

type Event struct {
	ID          string
	Complete    bool
	Details     interface{}
	RealTime    time.Time
	VirtualTime time.Time
}

func (e *Event) CompleteEvent() {
	e.Complete = true
}

func (e *Event) SetRealTime(t time.Time) {
	e.RealTime = t
}

func (e *Event) SetVirtualTime(t time.Time) {
	e.VirtualTime = t
}
