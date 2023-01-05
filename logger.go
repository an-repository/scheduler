/*
------------------------------------------------------------------------------------------------------------------------
####### Copyright (c) 2022-2023 Archivage Numérique.
####### All rights reserved.
####### Use of this source code is governed by a MIT style license that can be found in the LICENSE file.
------------------------------------------------------------------------------------------------------------------------
*/

package scheduler

import (
	"github.com/an-repository/zombie"
	"github.com/robfig/cron/v3"
)

type (
	Logger interface {
		zombie.Logger
		Trace(msg string, kv ...any)
		Error(err error, msg string, kv ...any)
	}

	cronLogger struct {
		logger Logger
	}
)

func newCronLogger(logger Logger) cron.Logger {
	return &cronLogger{logger}
}

func (cl *cronLogger) Info(msg string, kv ...any) {
	if cl.logger != nil {
		cl.logger.Trace("[scheduler] "+msg, kv...) //:::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
	}
}

func (cl *cronLogger) Error(err error, msg string, kv ...any) {
	if cl.logger != nil {
		cl.logger.Error(err, "[scheduler] "+msg, kv...) //::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
	}
}

/*
######################################################################################################## @(^_^)@ #######
*/
