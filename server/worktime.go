package main

import "time"

// schedule is when the piggy is "at work": its needs only drop then, so a
// weekend or a night without anybody around never kills it.
type schedule struct {
	loc         *time.Location
	startHour   int
	endHour     int
	daysToDeath int
}

// Longer gaps (the plugin disabled for months) are clamped: by then every need is
// long at zero anyway, and the day loop below stays short.
const maxElapsed = 60 * 24 * time.Hour

func (s schedule) workdayHours() float64 {
	return float64(s.endHour - s.startHour)
}

func isWorkday(t time.Time) bool {
	return t.Weekday() != time.Saturday && t.Weekday() != time.Sunday
}

func (s schedule) isWorkingTime(t time.Time) bool {
	t = t.In(s.loc)
	return isWorkday(t) && t.Hour() >= s.startHour && t.Hour() < s.endHour
}

// workingDuration is how much of [from, to) falls inside working hours.
func (s schedule) workingDuration(from, to time.Time) time.Duration {
	if !to.After(from) {
		return 0
	}
	if to.Sub(from) > maxElapsed {
		from = to.Add(-maxElapsed)
	}
	from, to = from.In(s.loc), to.In(s.loc)
	var total time.Duration
	for day := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, s.loc); day.Before(to); day = day.AddDate(0, 0, 1) {
		if !isWorkday(day) {
			continue
		}
		// time.Date instead of adding hours, so a daylight saving change keeps the hours on the clock.
		start := time.Date(day.Year(), day.Month(), day.Day(), s.startHour, 0, 0, 0, s.loc)
		end := time.Date(day.Year(), day.Month(), day.Day(), s.endHour, 0, 0, 0, s.loc)
		if from.After(start) {
			start = from
		}
		if to.Before(end) {
			end = to
		}
		if end.After(start) {
			total += end.Sub(start)
		}
	}
	return total
}
