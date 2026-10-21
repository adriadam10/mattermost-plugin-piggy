package main

import "testing"

func TestSettingsDefaults(t *testing.T) {
	s, err := configuration{}.settings()
	if err != nil {
		t.Fatal(err)
	}
	if s.name != "Bacon" || s.trigger != "piggy" || s.schedule.daysToDeath != 3 ||
		s.schedule.startHour != 8 || s.schedule.endHour != 18 || s.schedule.loc.String() != "UTC" || s.announceTeam != "" {
		t.Errorf("defaults: %+v", s)
	}
}

func TestSettingsValues(t *testing.T) {
	s, err := configuration{
		PetName: "Peppa", CommandTrigger: "/cerdito", DaysToDeath: "5", WorkdayStart: float64(9),
		WorkdayEnd: "17", TimeZone: "Europe/Madrid", AnnouncementChannel: "team/town-square",
	}.settings()
	if err != nil {
		t.Fatal(err)
	}
	if s.name != "Peppa" || s.trigger != "cerdito" || s.schedule.daysToDeath != 5 || s.schedule.workdayHours() != 8 ||
		s.announceTeam != "team" || s.announceChannel != "town-square" {
		t.Errorf("values: %+v", s)
	}
}

func TestSettingsInvalid(t *testing.T) {
	for _, c := range []configuration{
		{DaysToDeath: "0"},
		{DaysToDeath: "a lot"},
		{WorkdayStart: "18", WorkdayEnd: "8"},
		{TimeZone: "Mars/Olympus"},
		{CommandTrigger: "two words"},
		{AnnouncementChannel: "no-team"},
	} {
		if _, err := c.settings(); err == nil {
			t.Errorf("accepted %+v", c)
		}
	}
}

func TestPetName(t *testing.T) {
	for generation, want := range map[int]string{1: "Bacon", 2: "Bacon II", 4: "Bacon IV", 14: "Bacon XIV"} {
		if got := petName("Bacon", generation); got != want {
			t.Errorf("generation %d: %q", generation, got)
		}
	}
}

func TestMessagesTranslated(t *testing.T) {
	for key := range messages["en"] {
		if messages["es"][key] == "" {
			t.Errorf("%q has no Spanish text", key)
		}
	}
	for _, refusal := range []error{errDead, errAlive, errAsleep, errFull, errTired, errBadInput, errCooldown, errAwake} {
		if messages["en"]["err_"+refusal.Error()] == "" {
			t.Errorf("refusal %q has no message", refusal)
		}
	}
}

func TestAvatar(t *testing.T) {
	if _, err := avatarPNG(); err != nil {
		t.Fatal(err)
	}
}
