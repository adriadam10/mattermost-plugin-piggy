package main

import (
	"errors"
	"testing"
	"time"
)

func testSchedule(t *testing.T) schedule {
	t.Helper()
	loc, err := time.LoadLocation("Europe/Madrid")
	if err != nil {
		t.Fatal(err)
	}
	return schedule{loc: loc, startHour: 8, endHour: 18, daysToDeath: 3}
}

func at(s schedule, value string) time.Time {
	t, err := time.ParseInLocation("2006-01-02 15:04", value, s.loc)
	if err != nil {
		panic(err)
	}
	return t
}

func TestWorkingDuration(t *testing.T) {
	s := testSchedule(t)
	cases := []struct {
		from, to string
		want     time.Duration
	}{
		{"2026-10-19 07:00", "2026-10-19 09:30", 90 * time.Minute}, // Monday, from before opening
		{"2026-10-19 17:00", "2026-10-20 09:00", 2 * time.Hour},    // overnight
		{"2026-10-23 17:00", "2026-10-26 09:00", 2 * time.Hour},    // over the weekend
		{"2026-10-24 10:00", "2026-10-25 12:00", 0},                // Saturday to Sunday
		{"2026-10-19 08:00", "2026-10-26 08:00", 50 * time.Hour},   // a whole week
		{"2026-10-20 10:00", "2026-10-20 09:00", 0},                // backwards
	}
	for _, c := range cases {
		if got := s.workingDuration(at(s, c.from), at(s, c.to)); got != c.want {
			t.Errorf("%s → %s: got %v, want %v", c.from, c.to, got, c.want)
		}
	}
}

func TestDiesAfterDaysToDeath(t *testing.T) {
	s := testSchedule(t)
	born := at(s, "2026-10-19 08:00") // Monday
	pt := newPet(1, born)

	// Two and a half working days in, it is ill but still alive.
	pt.advance(at(s, "2026-10-21 13:00"), s)
	if pt.dead() || pt.Health >= sickHealth || pt.Fullness != 0 {
		t.Fatalf("after 25 working hours: %+v", pt)
	}
	// Three working days of 10 hours end on Wednesday evening.
	pt.advance(at(s, "2026-10-21 18:00"), s)
	if !pt.dead() {
		t.Fatalf("after 30 working hours it should be dead: %+v", pt)
	}
	if got := pt.mood(at(s, "2026-10-22 10:00"), s); got != moodDead {
		t.Errorf("mood = %s", got)
	}
}

func TestWeekendDoesNotHurt(t *testing.T) {
	s := testSchedule(t)
	pt := newPet(1, at(s, "2026-10-23 18:00")) // Friday evening
	pt.advance(at(s, "2026-10-26 08:00"), s)
	if pt.Fullness != maxStat || pt.Health != maxStat {
		t.Errorf("a weekend changed the needs: %+v", pt)
	}
}

func TestCareKeepsItAlive(t *testing.T) {
	s := testSchedule(t)
	pt := newPet(1, at(s, "2026-10-19 08:00"))
	day := at(s, "2026-10-19 00:00")
	// Four weeks of two people dropping by in the morning and after lunch.
	for range 28 {
		for _, hour := range []time.Duration{9, 14} {
			now := day.Add(hour * time.Hour)
			pt.advance(now, s)
			for _, user := range []string{"u1", "u2"} {
				for _, a := range []string{actionFeed, actionPet, actionPlay} {
					_ = pt.apply(a, user, now, s)
				}
			}
		}
		day = day.AddDate(0, 0, 1)
	}
	if pt.dead() || pt.Health < recoveredHealth {
		t.Errorf("a cared for piggy is unwell: %+v", pt)
	}
}

func TestApplyRefusals(t *testing.T) {
	s := testSchedule(t)
	now := at(s, "2026-10-19 10:00")
	pt := newPet(1, now)
	if err := pt.apply(actionFeed, "u1", now, s); !errors.Is(err, errFull) {
		t.Errorf("feeding a full piggy: %v", err)
	}
	if err := pt.apply(actionAdopt, "u1", now, s); !errors.Is(err, errAlive) {
		t.Errorf("adopting a live piggy: %v", err)
	}
	if err := pt.apply(actionSleep, "u1", now, s); !errors.Is(err, errAwake) {
		t.Errorf("sending a rested piggy to bed: %v", err)
	}
	pt.Energy = 30
	if err := pt.apply(actionSleep, "u3", now, s); err != nil {
		t.Fatal(err)
	}
	if got := pt.mood(now, s); got != moodSleeping {
		t.Errorf("mood = %s", got)
	}
	if err := pt.apply(actionSleep, "u2", now, s); !errors.Is(err, errAsleep) {
		t.Errorf("sending a sleeping piggy to bed: %v", err)
	}
	if err := pt.apply(actionPet, "u2", now, s); err != nil || pt.asleep(now) {
		t.Errorf("petting should wake it up: %v, sleep until %d", err, pt.SleepUntil)
	}
	if err := pt.apply("dance", "u1", now, s); !errors.Is(err, errBadInput) {
		t.Errorf("unknown action: %v", err)
	}
}

func TestAdoptAfterDeath(t *testing.T) {
	s := testSchedule(t)
	now := at(s, "2026-10-19 10:00")
	pt := newPet(3, now)
	pt.DiedAt = now.UnixMilli()
	if err := pt.apply(actionFeed, "u1", now, s); !errors.Is(err, errDead) {
		t.Errorf("feeding a dead piggy: %v", err)
	}
	if err := pt.apply(actionAdopt, "u1", now, s); err != nil {
		t.Fatal(err)
	}
	if pt.dead() || pt.Generation != 4 || pt.bestFriend() != "u1" {
		t.Errorf("adopted: %+v", pt)
	}
}

func TestNapRefillsEnergy(t *testing.T) {
	s := testSchedule(t)
	now := at(s, "2026-10-19 10:00")
	pt := newPet(1, now)
	pt.Energy = 10
	_ = pt.apply(actionSleep, "u1", now, s)
	pt.advance(now.Add(napLength/2), s)
	if pt.Energy < 55 || pt.Energy > 60 {
		t.Errorf("half a nap: energy %v", pt.Energy)
	}
	pt.advance(now.Add(2*time.Hour), s)
	if pt.asleep(now.Add(2*time.Hour)) || pt.Energy < 90 {
		t.Errorf("after the nap: %+v", pt)
	}
}

func TestBestFriendAndLog(t *testing.T) {
	now := time.Unix(0, 0)
	pt := newPet(1, now)
	for range 10 {
		pt.record(actionPet, "u2", now)
	}
	pt.record(actionPet, "u1", now)
	if pt.bestFriend() != "u2" || len(pt.Log) != logLength || pt.Log[0].UserID != "u1" {
		t.Errorf("best friend %q, log %+v", pt.bestFriend(), pt.Log)
	}
}

func TestCooldownPerUserAndAction(t *testing.T) {
	s := testSchedule(t)
	now := at(s, "2026-10-19 10:00")
	pt := newPet(1, now)
	pt.Happiness = 10
	if err := pt.apply(actionPet, "u1", now, s); err != nil {
		t.Fatal(err)
	}
	if err := pt.apply(actionPet, "u1", now.Add(time.Minute), s); !errors.Is(err, errCooldown) {
		t.Errorf("petting twice in a row: %v", err)
	}
	// Somebody else, or another action, is fine.
	if err := pt.apply(actionPet, "u2", now.Add(time.Minute), s); err != nil {
		t.Errorf("another user: %v", err)
	}
	if err := pt.apply(actionPlay, "u1", now.Add(time.Minute), s); err != nil {
		t.Errorf("another action: %v", err)
	}
	if got := pt.cooldowns("u1", now.Add(time.Minute)); len(got) != 2 || got[actionPet] != now.Add(actionCooldown).UnixMilli() {
		t.Errorf("cooldowns: %v", got)
	}
	later := now.Add(actionCooldown + time.Minute)
	if err := pt.apply(actionPet, "u1", later, s); err != nil {
		t.Errorf("after the cooldown: %v", err)
	}
	// Old entries are dropped when anybody acts.
	if _, ok := pt.LastAction["u2"]; ok {
		t.Errorf("u2's finished cooldown is still stored: %v", pt.LastAction)
	}
}

func TestNightWake(t *testing.T) {
	s := testSchedule(t)
	night := at(s, "2026-10-19 21:00")
	pt := newPet(1, night)
	pt.Happiness = 50
	if got := pt.mood(night, s); got != moodNight {
		t.Fatalf("mood at night = %s", got)
	}
	if err := pt.apply(actionPet, "u1", night, s); err != nil {
		t.Fatal(err)
	}
	if got := pt.mood(night.Add(time.Minute), s); got == moodNight {
		t.Errorf("still asleep after petting")
	}
	if got := pt.mood(night.Add(nightWake+time.Minute), s); got != moodNight {
		t.Errorf("did not doze off again: %s", got)
	}
	// A refused action does not count as waking it: the caller drops the change.
	full := newPet(1, night)
	if err := full.apply(actionFeed, "u1", night, s); !errors.Is(err, errFull) {
		t.Errorf("feeding a full piggy: %v", err)
	}
}
