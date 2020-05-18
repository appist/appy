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
func (l *logger) Debug(args ...interface{}) {
	l.Logger.Debug(args...)
}

// Info uses fmt.Sprintf to log a templated information message.
func (l *logger) Info(args ...interface{}) {
	if len(args) > 0 {
		info := args[0].(string)

		if skipMessageRegex.Match([]byte(info)) || info == "" {
			return
		}

		if startMessageRegex.Match([]byte(info)) {
			for _, info := range l.worker.Info() {
				l.Logger.Info(info)
			}

			return
		}

		if strings.Contains(info, "Starting graceful shutdown") {
			l.Logger.Info("* Gracefully shutting down the worker...")
			return
		}
	}

	l.Logger.Info(args...)
}

// Warn uses fmt.Sprintf to log a templated warning message.
func (l *logger) Warn(args ...interface{}) {
	l.Logger.Warn(args...)
}

// Error uses fmt.Sprintf to log a templated error message.
func (l *logger) Error(args ...interface{}) {
	l.Logger.Error(args...)
}

// Fatal uses fmt.Sprintf to log a templated fatal message.
func (l *logger) Fatal(args ...interface{}) {
	l.Logger.Fatal(args...)
}
