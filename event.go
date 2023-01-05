/*
------------------------------------------------------------------------------------------------------------------------
####### Copyright (c) 2022-2023 Archivage Numérique.
####### All rights reserved.
####### Use of this source code is governed by a MIT style license that can be found in the LICENSE file.
------------------------------------------------------------------------------------------------------------------------
*/

package scheduler

import "github.com/robfig/cron/v3"

type event struct {
	name     string
	disabled bool
	data     string
	filter   func(*event)
	callback Callback
}

func (e *event) Run() {
	e.filter(e)
}

type runOnce struct {
	event    *event
	entryID  cron.EntryID
	schedule cron.Schedule
	cron     *cron.Cron
}

func (ro *runOnce) Run() {
	ro.cron.Remove(ro.entryID)

	if ro.schedule != nil {
		ro.cron.Schedule(ro.schedule, ro.event)
	}

	ro.event.Run()
}

/*
######################################################################################################## @(^_^)@ #######
*/
