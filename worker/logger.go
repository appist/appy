package worker

import (
	"regexp"
	"strings"

	"github.com/appist/appy/support"
)

var (
	skipMessageRegex  = regexp.MustCompile(`(?i)(bye!| done| shutting down...|waiting for all workers to finish...|all workers have finished|send signal tstp|starting processing)`)
	startMessageRegex = regexp.MustCompile(`(?i)(send signal term)`)
)

type logger struct {
	*support.Logger
	worker *Engine
}

// Debug uses fmt.Sprintf to log a templated debug message.
func (l *logger) Debug(format string, args ...interface{}) {
	l.Logger.Debugf(format, args...)
}

// Info uses fmt.Sprintf to log a templated information message.
func (l *logger) Info(format string, args ...interface{}) {
	if skipMessageRegex.Match([]byte(format)) || format == "" {
		return
	}

	if startMessageRegex.Match([]byte(format)) {
		for _, info := range l.worker.Info() {
			l.Logger.Info(info)
		}

		return
	}

	if strings.Contains(format, "Starting graceful shutdown") {
		l.Logger.Info("* Gracefully shutting down the worker...")
		return
	}

	l.Logger.Infof(format, args...)
}

// Warn uses fmt.Sprintf to log a templated warning message.
func (l *logger) Warn(format string, args ...interface{}) {
	l.Logger.Warnf(format, args...)
}

// Error uses fmt.Sprintf to log a templated error message.
func (l *logger) Error(format string, args ...interface{}) {
	l.Logger.Errorf(format, args...)
}

// Fatal uses fmt.Sprintf to log a templated fatal message.
func (l *logger) Fatal(format string, args ...interface{}) {
	l.Logger.Fatalf(format, args...)
}
