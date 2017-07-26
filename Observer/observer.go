package Observer

import (
	"fmt"
)

// Event generic interface for sending things across channels
type Event interface {
}

// MessageEvent for simple string messages
type MessageEvent struct {
	Message string
}

// Subject type to notify observers
type Subject struct {
	eventChannels map[string][]chan Event
}

// Observer basic observer type
type Observer struct {
	Chnl    chan Event
	Handler func(Event)
}

// Process waits on channel and passes event to handler
func (o *Observer) Process() {
	go func() {
		for n := range o.Chnl {
			o.Handler(n)
		}
	}()
}

// NewSubject creates a new subject
func NewSubject() *Subject {
	s := Subject{}
	s.eventChannels = make(map[string][]chan Event)
	return &s
}

// AddObserver adds channel to the list of observers for the event
func (s *Subject) AddObserver(event string) chan Event {
	chnl := make(chan Event)
	s.eventChannels[event] = append(s.eventChannels[event], chnl)
	return chnl
}

// NotifyObservers sends paylod to all of the observers
func (s *Subject) NotifyObservers(event string, payload Event) error {
	channelList, present := s.eventChannels[event]
	if !present {
		return fmt.Errorf("No channels for event: %s", event)
	}

	for _, chn := range channelList {
		chn <- payload
	}

	return nil
}
