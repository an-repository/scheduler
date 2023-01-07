/*
------------------------------------------------------------------------------------------------------------------------
####### Copyright (c) 2022-2023 Archivage Numérique.
####### All rights reserved.
####### Use of this source code is governed by a MIT style license that can be found in the LICENSE file.
------------------------------------------------------------------------------------------------------------------------
*/

package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/an-repository/errors"
	"github.com/an-repository/zombie"
	"github.com/robfig/cron/v3"
)

type (
	Scheduler struct {
		cron   *cron.Cron
		parser cron.Parser
		logger Logger
		events map[string]*event
		mutex  sync.RWMutex
		zombie *zombie.Zombie
	}

	Option func(*Scheduler)

	Callback func(name string, data string)

	Event struct {
		Name     string        `dm:"name"`
		Disabled bool          `dm:"disabled"`
		After    time.Duration `dm:"after"`
		Repeat   string        `dm:"repeat"`
		Data     string        `dm:"data"`
	}
)

func WithLogger(logger Logger) Option {
	return func(s *Scheduler) {
		s.logger = logger
	}
}

func New(opts ...Option) *Scheduler {
	s := &Scheduler{
		events: make(map[string]*event),
	}

	for _, option := range opts {
		option(s)
	}

	cl := newCronLogger(s.logger)

	s.parser = cron.NewParser(
		cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
	)

	s.cron = cron.New(
		cron.WithChain(
			cron.Recover(cl),
			cron.SkipIfStillRunning(cl),
		),
		cron.WithLogger(cl),
		cron.WithParser(s.parser),
	)

	return s
}

func (s *Scheduler) createEvent(e *Event, cb Callback) (*event, error) {
	if cb == nil {
		return nil, errors.New("sending function must not be nil") /////////////////////////////////////////////////////
	}

	if e == nil {
		return nil, errors.New("event must not be nil") ////////////////////////////////////////////////////////////////
	}

	if e.Name == "" {
		return nil, errors.New("event name must not be empty") /////////////////////////////////////////////////////////
	}

	if _, ok := s.events[e.Name]; ok {
		return nil, errors.New("this event name is already used", "name", e.Name) //////////////////////////////////////
	}

	event := &event{
		name:     e.Name,
		disabled: e.Disabled,
		data:     e.Data,
		filter:   s.sendEvent,
		callback: cb,
	}

	return event, nil
}

func (s *Scheduler) AddEventList(eList []*Event, cb Callback) error {
	for _, e := range eList {
		if err := s.AddEvent(e, cb); err != nil {
			return err
		}
	}

	return nil
}

func (s *Scheduler) AddEvent(e *Event, cb Callback) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	event, err := s.createEvent(e, cb)
	if err != nil {
		return err
	}

	var ro *runOnce

	if e.After != 0 {
		ro = &runOnce{
			event: event,
			cron:  s.cron,
		}

		entryID, err := s.cron.AddJob(fmt.Sprintf("@every %s", e.After.String()), ro)
		if err != nil {
			return errors.WithMessage( /////////////////////////////////////////////////////////////////////////////////
				err,
				"unable to add this event",
				"name", e.Name,
			)
		}

		ro.entryID = entryID
	}

	if e.Repeat == "" {
		if e.After != 0 {
			s.events[e.Name] = event
			return nil
		}

		return errors.New( /////////////////////////////////////////////////////////////////////////////////////////////
			"at least one of the fields After or Repeat must be specified",
			"name", e.Name,
		)
	}

	schedule, err := s.parser.Parse(e.Repeat)
	if err != nil {
		if ro != nil {
			s.cron.Remove(ro.entryID)
		}

		return errors.WithMessage( /////////////////////////////////////////////////////////////////////////////////////
			err,
			"unable to add this event (parser error)",
			"name", e.Name,
			"repeat", e.Repeat,
		)
	}

	if ro == nil {
		_ = s.cron.Schedule(schedule, event)
	} else {
		ro.schedule = schedule
	}

	s.events[e.Name] = event

	return nil
}

func (s *Scheduler) sendEvent(e *event) {
	var disabled bool

	s.mutex.RLock()
	disabled = e.disabled
	s.mutex.RUnlock()

	if disabled {
		return
	}

	e.callback(e.name, e.data)
}

func (s *Scheduler) Disable(name string, state bool) error {
	var (
		e  *event
		ok bool
	)

	s.mutex.RLock()
	e, ok = s.events[name]
	if ok {
		e.disabled = state
	}
	s.mutex.RUnlock()

	if !ok {
		return errors.New("this event doesn't exist", "name", name) ////////////////////////////////////////////////////
	}

	return nil
}

func (s *Scheduler) FireEvent(name string) error {
	var (
		e  *event
		ok bool
	)

	s.mutex.RLock()
	e, ok = s.events[name]
	s.mutex.RUnlock()

	if !ok {
		return errors.New("this event doesn't exist", "name", name) ////////////////////////////////////////////////////
	}

	e.callback(e.name, e.data)

	return nil
}

func (s *Scheduler) Start() error {
	if s.zombie != nil {
		return errors.New("scheduler already started") /////////////////////////////////////////////////////////////////
	}

	s.zombie = zombie.Go(
		context.Background(),
		func(_ context.Context, _ *zombie.Zombie) error {
			s.cron.Run()
			return nil
		},
		zombie.WithName("scheduler"),
		zombie.WithLogger(s.logger),
	)

	return nil
}

func (s *Scheduler) Stop() error {
	if s.zombie == nil {
		return errors.New("scheduler not started") /////////////////////////////////////////////////////////////////////
	}

	<-s.cron.Stop().Done()
	s.zombie.Wait()

	s.zombie = nil

	return nil
}

/*
######################################################################################################## @(^_^)@ #######
*/
