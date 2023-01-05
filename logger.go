/*
------------------------------------------------------------------------------------------------------------------------
####### Copyright (c) 2022-2023 Archivage Numérique.
####### All rights reserved.
####### Use of this source code is governed by a MIT style license that can be found in the LICENSE file.
------------------------------------------------------------------------------------------------------------------------
*/

package scheduler

import (
	"github.com/an-repository/logger"
	"github.com/robfig/cron/v3"
)

type cronLogger struct {
	logger *logger.Logger
}

func newCronLogger(logger *logger.Logger) cron.Logger {
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
