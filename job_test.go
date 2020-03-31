package appy

import (
	"testing"
	"time"
)

type JobSuite struct {
	TestSuite
}

func (s *JobSuite) TestParseJobOptions() {
	opts := parseJobOptions(nil)
	s.Equal(0, len(opts))

	deadline := time.Now().Add(5 * time.Minute)
	opts = parseJobOptions(&JobOptions{
		Deadline:  deadline,
		MaxRetry:  10,
		Queue:     "critical",
		Timeout:   10 * time.Second,
		UniqueTTL: 10 * time.Second,
	})
	s.Equal(5, len(opts))
}

func TestJobSuite(t *testing.T) {
	RunTestSuite(t, new(JobSuite))
}
