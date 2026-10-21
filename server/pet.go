package main

import (
	"errors"
	"time"
)

// Actions anybody can do, from the widget or the slash command.
const (
	actionFeed  = "feed"
	actionPet   = "pet"
	actionPlay  = "play"
	actionSleep = "sleep"
	actionAdopt = "adopt"
)

// Moods, from the most to the least important; the webapp draws one of them.
const (
	moodDead     = "dead"
	moodSleeping = "sleeping"
	moodNight    = "night"
	moodSick     = "sick"
	moodHungry   = "hungry"
	moodSad      = "sad"
	moodTired    = "tired"
	moodHappy    = "happy"
	moodOK       = "ok"
)

var (
	errDead     = errors.New("dead")
	errAlive    = errors.New("alive")
	errAsleep   = errors.New("asleep")
	errFull     = errors.New("full")
	errTired    = errors.New("tired")
	errBadInput = errors.New("unknown")
	errCooldown = errors.New("cooldown")
	errAwake    = errors.New("not_sleepy")
)

const (
	maxStat = 100.0
	// Below this the piggy looks sick and the channel is told; above recoveredHealth it is told again next time.
	sickHealth      = 40.0
	recoveredHealth = 60.0
	lowStat         = 25.0
	// Above this energy it refuses to go to bed: with everybody pressing every
	// button, sending it to sleep kept it lying down all afternoon.
	sleepyEnergy = 50.0
	// What a game costs: with half a dozen people playing every cooldown it
	// still lasts the afternoon.
	playEnergy = 5.0
	// A short nap, since any action wakes the piggy up anyway.
	napLength = 20 * time.Minute
	// How long an action keeps it awake out of working hours before it dozes off again.
	nightWake = 15 * time.Minute
	// Each person can repeat an action only this often, so mashing a button does not
	// fill the bars on its own: keeping the piggy well takes the whole team.
	actionCooldown = 15 * time.Minute
	logLength      = 8
)

type event struct {
	UserID string `json:"user_id"`
	Action string `json:"action"`
	At     int64  `json:"at"`
}

// pet is the shared piggy as stored in the KV store. Stats go from 0 to 100.
type pet struct {
	Generation int     `json:"generation"`
	BornAt     int64   `json:"born_at"`
	DiedAt     int64   `json:"died_at,omitempty"`
	Fullness   float64 `json:"fullness"`
	Happiness  float64 `json:"happiness"`
	Energy     float64 `json:"energy"`
	Health     float64 `json:"health"`
	SleepUntil int64   `json:"sleep_until,omitempty"`
	// Out of working hours it sleeps unless somebody woke it up until then.
	AwakeUntil int64 `json:"awake_until,omitempty"`
	// Last time the needs were brought up to date.
	UpdatedAt     int64          `json:"updated_at"`
	SickAnnounced bool           `json:"sick_announced,omitempty"`
	Log           []event        `json:"log"`
	CareCount     map[string]int `json:"care_count"`
	// When each user last did each action, for the cooldown.
	LastAction map[string]map[string]int64 `json:"last_action,omitempty"`
}

func newPet(generation int, now time.Time) pet {
	return pet{
		Generation: generation,
		BornAt:     now.UnixMilli(),
		Fullness:   maxStat,
		Happiness:  maxStat,
		Energy:     maxStat,
		Health:     maxStat,
		UpdatedAt:  now.UnixMilli(),
		Log:        []event{},
		CareCount:  map[string]int{},
	}
}

func clamp(v float64) float64 {
	return min(maxStat, max(0, v))
}

func (p *pet) dead() bool {
	return p.DiedAt != 0
}

func (p *pet) asleep(now time.Time) bool {
	return p.SleepUntil > now.UnixMilli()
}

// rates per working hour, derived from the days to death: with nobody around the
// piggy is hungry after a third of them, and starving takes the rest to kill it.
// Happiness and energy last longer, so a couple of visits a day keep it well.
type rates struct {
	fullness, happiness, energy, health, recovery float64
}

func (s schedule) rates() rates {
	total := float64(s.daysToDeath) * s.workdayHours()
	return rates{
		fullness:  maxStat / (total / 3),
		happiness: maxStat / (total * 2 / 3),
		energy:    maxStat / total,
		health:    maxStat / (total * 2 / 3),
		recovery:  maxStat / (total / 3),
	}
}

// hoursAtZero is how long a stat that started at value and dropped at rate
// spent at zero during the last hours.
func hoursAtZero(value, rate, hours float64) float64 {
	return max(0, hours-value/rate)
}

// advance brings the needs up to now. It is safe to call as often as wanted.
func (p *pet) advance(now time.Time, s schedule) {
	if p.dead() {
		return
	}
	last := time.UnixMilli(p.UpdatedAt)
	p.UpdatedAt = now.UnixMilli()
	if !now.After(last) {
		return
	}

	// A nap refills energy in real time, also at night.
	if p.SleepUntil > last.UnixMilli() {
		slept := time.UnixMilli(min(p.SleepUntil, now.UnixMilli())).Sub(last)
		p.Energy = clamp(p.Energy + maxStat*slept.Hours()/napLength.Hours())
	}

	hours := s.workingDuration(last, now).Hours()
	if hours == 0 {
		return
	}
	r := s.rates()
	starving := max(hoursAtZero(p.Fullness, r.fullness, hours), hoursAtZero(p.Happiness, r.happiness, hours))
	wellFed := p.Fullness > 50 && p.Happiness > 50
	p.Fullness = clamp(p.Fullness - r.fullness*hours)
	p.Happiness = clamp(p.Happiness - r.happiness*hours)
	// Energy only drops while awake.
	awakeFrom := time.UnixMilli(max(last.UnixMilli(), p.SleepUntil))
	p.Energy = clamp(p.Energy - r.energy*s.workingDuration(awakeFrom, now).Hours())
	switch {
	case starving > 0:
		p.Health = clamp(p.Health - r.health*starving)
	case wellFed:
		p.Health = clamp(p.Health + r.recovery*hours)
	}
	if p.Health == 0 {
		p.DiedAt = now.UnixMilli()
		p.SleepUntil = 0
		return
	}
	// Worn out, it falls asleep on its own.
	if p.Energy == 0 && !p.asleep(now) {
		p.SleepUntil = now.Add(napLength).UnixMilli()
	}
}

func (p *pet) mood(now time.Time, s schedule) string {
	switch {
	case p.dead():
		return moodDead
	case p.asleep(now):
		return moodSleeping
	case p.nightSleep(now, s):
		return moodNight
	case p.Health < sickHealth:
		return moodSick
	case p.Fullness < lowStat:
		return moodHungry
	case p.Happiness < lowStat:
		return moodSad
	case p.Energy < lowStat:
		return moodTired
	case p.Happiness > 75 && p.Fullness > 50:
		return moodHappy
	default:
		return moodOK
	}
}

// nightSleep is the sleep out of working hours, which an action interrupts for a while.
func (p *pet) nightSleep(now time.Time, s schedule) bool {
	return !s.isWorkingTime(now) && p.AwakeUntil <= now.UnixMilli()
}

// wake ends a nap, keeping the energy it already gave, and at night keeps the
// piggy up for a while.
func (p *pet) wake(now time.Time, s schedule) {
	p.SleepUntil = 0
	if !s.isWorkingTime(now) {
		p.AwakeUntil = now.Add(nightWake).UnixMilli()
	}
}

// apply does an action of userID on the piggy, already advanced to now.
func (p *pet) apply(action, userID string, now time.Time, s schedule) error {
	if action == actionAdopt {
		if !p.dead() {
			return errAlive
		}
		*p = newPet(p.Generation+1, now)
		p.record(action, userID, now)
		return nil
	}
	if p.dead() {
		return errDead
	}
	if _, waiting := p.cooldowns(userID, now)[action]; waiting {
		return errCooldown
	}
	if action == actionSleep {
		if p.asleep(now) || p.nightSleep(now, s) {
			return errAsleep
		}
		if p.Energy > sleepyEnergy {
			return errAwake
		}
		p.SleepUntil = now.Add(napLength).UnixMilli()
		p.AwakeUntil = 0
		p.record(action, userID, now)
		return nil
	}
	// Feeding, petting and playing wake it up. If the action is refused the
	// whole change is dropped, so it does not wake up for nothing.
	if p.asleep(now) || p.nightSleep(now, s) {
		p.wake(now, s)
	}
	switch action {
	case actionFeed:
		if p.Fullness >= 95 {
			return errFull
		}
		p.Fullness = clamp(p.Fullness + 35)
		p.Happiness = clamp(p.Happiness + 5)
	case actionPet:
		p.Happiness = clamp(p.Happiness + 10)
	case actionPlay:
		if p.Energy < 15 {
			return errTired
		}
		p.Happiness = clamp(p.Happiness + 25)
		p.Energy = clamp(p.Energy - playEnergy)
		p.Fullness = clamp(p.Fullness - 5)
	default:
		return errBadInput
	}
	p.record(action, userID, now)
	return nil
}

func (p *pet) record(action, userID string, now time.Time) {
	p.Log = append([]event{{UserID: userID, Action: action, At: now.UnixMilli()}}, p.Log...)
	if len(p.Log) > logLength {
		p.Log = p.Log[:logLength]
	}
	if p.CareCount == nil {
		p.CareCount = map[string]int{}
	}
	p.CareCount[userID]++
	if action == actionAdopt {
		return
	}
	if p.LastAction == nil {
		p.LastAction = map[string]map[string]int64{}
	}
	p.forgetCooldowns(now)
	if p.LastAction[userID] == nil {
		p.LastAction[userID] = map[string]int64{}
	}
	p.LastAction[userID][action] = now.UnixMilli()
}

// cooldowns is when each action userID is still waiting for becomes available again.
func (p *pet) cooldowns(userID string, now time.Time) map[string]int64 {
	ready := map[string]int64{}
	for action, at := range p.LastAction[userID] {
		if until := at + actionCooldown.Milliseconds(); until > now.UnixMilli() {
			ready[action] = until
		}
	}
	return ready
}

// forgetCooldowns drops the ones that are over, so the stored piggy does not grow.
func (p *pet) forgetCooldowns(now time.Time) {
	for userID, actions := range p.LastAction {
		for action, at := range actions {
			if at+actionCooldown.Milliseconds() <= now.UnixMilli() {
				delete(actions, action)
			}
		}
		if len(actions) == 0 {
			delete(p.LastAction, userID)
		}
	}
}

// bestFriend is who looked after this piggy the most.
func (p *pet) bestFriend() string {
	best, most := "", 0
	for userID, n := range p.CareCount {
		if n > most || (n == most && userID < best) {
			best, most = userID, n
		}
	}
	return best
}
