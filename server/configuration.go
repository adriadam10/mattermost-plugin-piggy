package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// configuration mirrors settings_schema in plugin.json. The System Console can
// store a number setting as a JSON number or as a string, so those are decoded loosely.
type configuration struct {
	PetName             string
	CommandTrigger      string
	DaysToDeath         any
	WorkdayStart        any
	WorkdayEnd          any
	TimeZone            string
	AnnouncementChannel string
}

// settings is the validated configuration.
type settings struct {
	name     string
	trigger  string
	schedule schedule
	// Empty when announcements are off.
	announceTeam, announceChannel string
}

func number(name string, value any, def, lo, hi int) (int, error) {
	var n int
	switch v := value.(type) {
	case nil:
		return def, nil
	case float64:
		n = int(v)
	case string:
		if v == "" {
			return def, nil
		}
		parsed, err := strconv.Atoi(v)
		if err != nil {
			return 0, fmt.Errorf("%s must be a whole number, got %q", name, v)
		}
		n = parsed
	default:
		return 0, fmt.Errorf("%s has an unexpected type %T", name, v)
	}
	if n < lo || n > hi {
		return 0, fmt.Errorf("%s must be between %d and %d, got %d", name, lo, hi, n)
	}
	return n, nil
}

func orDefault(value, def string) string {
	if v := strings.TrimSpace(value); v != "" {
		return v
	}
	return def
}

func (c configuration) settings() (*settings, error) {
	s := &settings{
		name:    orDefault(c.PetName, "Bacon"),
		trigger: strings.TrimPrefix(orDefault(c.CommandTrigger, "piggy"), "/"),
	}
	if strings.ContainsAny(s.trigger, " /") {
		return nil, fmt.Errorf("CommandTrigger must be a single word, got %q", s.trigger)
	}
	var err error
	if s.schedule.daysToDeath, err = number("DaysToDeath", c.DaysToDeath, 3, 1, 30); err != nil {
		return nil, err
	}
	if s.schedule.startHour, err = number("WorkdayStart", c.WorkdayStart, 8, 0, 23); err != nil {
		return nil, err
	}
	if s.schedule.endHour, err = number("WorkdayEnd", c.WorkdayEnd, 18, 1, 24); err != nil {
		return nil, err
	}
	if s.schedule.endHour <= s.schedule.startHour {
		return nil, fmt.Errorf("WorkdayEnd (%d) must be after WorkdayStart (%d)", s.schedule.endHour, s.schedule.startHour)
	}
	if s.schedule.loc, err = time.LoadLocation(orDefault(c.TimeZone, "UTC")); err != nil {
		return nil, fmt.Errorf("TimeZone: %w", err)
	}
	if channel := strings.TrimSpace(c.AnnouncementChannel); channel != "" {
		team, name, ok := strings.Cut(channel, "/")
		if !ok || team == "" || name == "" {
			return nil, fmt.Errorf("AnnouncementChannel must look like team-name/channel-name, got %q", channel)
		}
		s.announceTeam, s.announceChannel = team, name
	}
	return s, nil
}

func (p *Plugin) OnConfigurationChange() error {
	var c configuration
	if err := p.API.LoadPluginConfiguration(&c); err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}
	s, err := c.settings()
	if err != nil {
		return err
	}
	old := p.settings.Swap(s)
	// On the first load the plugin is not active yet; OnActivate does the rest.
	if old == nil || p.botID == "" {
		return nil
	}
	return p.applySettings(old, s)
}
