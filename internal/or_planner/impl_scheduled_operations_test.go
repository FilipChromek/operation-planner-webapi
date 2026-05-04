package or_planner

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type ScheduleSuite struct {
	suite.Suite
}

func TestScheduleSuite(t *testing.T) {
	suite.Run(t, new(ScheduleSuite))
}

func (s *ScheduleSuite) Test_Reconcile_OrdersByStart() {
	now := time.Now()
	ops := []ScheduledOperation{
		{Id: "b", ScheduledStart: now.Add(2 * time.Hour), DurationMinutes: 60},
		{Id: "a", ScheduledStart: now, DurationMinutes: 60},
	}
	out := reconcileSchedule(ops)
	s.Equal("a", out[0].Id)
	s.Equal("b", out[1].Id)
}

func (s *ScheduleSuite) Test_Reconcile_ResolvesOverlap() {
	now := time.Now()
	ops := []ScheduledOperation{
		{Id: "a", ScheduledStart: now, DurationMinutes: 60},
		{Id: "b", ScheduledStart: now.Add(30 * time.Minute), DurationMinutes: 60},
	}
	out := reconcileSchedule(ops)
	expectedB := now.Add(60 * time.Minute)
	s.WithinDuration(expectedB, out[1].ScheduledStart, time.Second)
}
