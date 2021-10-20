package main

import (
	"github.com/phuslu/log"
	"os"
)

// LogAdapter struct
type LogAdapter struct {
	logger *log.Logger
}

// Errorf func
func (a *LogAdapter) Errorf(s string, i ...interface{}) {
	a.logger.Error().Msgf(s, i...)
}

// Warningf func
func (a *LogAdapter) Warningf(s string, i ...interface{}) {
	a.logger.Warn().Msgf(s, i...)
}

// Infof func
func (a *LogAdapter) Infof(s string, i ...interface{}) {
	a.logger.Info().Msgf(s, i...)
}

// Debugf func
func (a *LogAdapter) Debugf(s string, i ...interface{}) {
	a.logger.Debug().Msgf(s, i...)
}

func initLog(verbose bool) {
	w := &log.ConsoleWriter{
		//ColorOutput:    true,
		QuoteString:    false,
		EndWithMessage: false,
	}
	if log.IsTerminal(os.Stderr.Fd()) {
		w.ColorOutput = true
	}
	var lvl = log.InfoLevel
	if verbose {
		lvl = log.DebugLevel
	}
	log.DefaultLogger = log.Logger{
		Level:      lvl,
		TimeFormat: "2006.01.02 15:04:05",
		Caller:     0,
		Writer:     w,
	}
}
