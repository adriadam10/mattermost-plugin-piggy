package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

func (p *Plugin) newRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/pet", p.handlePet)
	mux.HandleFunc("POST /api/v1/action", p.handleAction)
	mux.HandleFunc("POST /api/v1/visibility", p.handleVisibility)
	return requireUser(mux)
}

func requireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Mattermost-User-Id") == "" {
			http.Error(w, "not authorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type logEntry struct {
	Username string `json:"username"`
	Action   string `json:"action"`
	At       int64  `json:"at"`
}

// petView is the piggy as the webapp draws it.
type petView struct {
	Name       string     `json:"name"`
	Generation int        `json:"generation"`
	BornAt     int64      `json:"born_at"`
	DiedAt     int64      `json:"died_at"`
	Fullness   float64    `json:"fullness"`
	Happiness  float64    `json:"happiness"`
	Energy     float64    `json:"energy"`
	Health     float64    `json:"health"`
	SleepUntil int64      `json:"sleep_until"`
	Mood       string     `json:"mood"`
	BestFriend string     `json:"best_friend"`
	Log        []logEntry `json:"log"`
	Trigger    string     `json:"trigger"`
	// Only for the user asking: when each action they are waiting for is available again.
	Cooldowns map[string]int64 `json:"cooldowns,omitempty"`
}

func (p *Plugin) username(userID string) string {
	if userID == "" {
		return ""
	}
	user, appErr := p.API.GetUser(userID)
	if appErr != nil {
		return ""
	}
	return user.Username
}

// view is the piggy for userID; with no user (a broadcast) it has no cooldowns.
func (p *Plugin) view(pt *pet, now time.Time, userID string) petView {
	s := p.settings.Load()
	v := petView{
		Name:       petName(s.name, pt.Generation),
		Generation: pt.Generation,
		BornAt:     pt.BornAt,
		DiedAt:     pt.DiedAt,
		Fullness:   pt.Fullness,
		Happiness:  pt.Happiness,
		Energy:     pt.Energy,
		Health:     pt.Health,
		SleepUntil: pt.SleepUntil,
		Mood:       pt.mood(now, s.schedule),
		BestFriend: p.username(pt.bestFriend()),
		Log:        make([]logEntry, 0, len(pt.Log)),
		Trigger:    s.trigger,
	}
	if userID != "" {
		v.Cooldowns = pt.cooldowns(userID, now)
	}
	for _, e := range pt.Log {
		v.Log = append(v.Log, logEntry{Username: p.username(e.UserID), Action: e.Action, At: e.At})
	}
	return v
}

func mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (p *Plugin) handlePet(w http.ResponseWriter, r *http.Request) {
	// Advancing on read keeps the stats fresh between ticks; it saves too.
	_, after, err := p.update(func(*pet, time.Time) error { return nil })
	if err != nil {
		p.API.LogError("Failed to load the piggy", "err", err.Error())
		http.Error(w, "could not load the piggy", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, p.view(&after, time.Now(), r.Header.Get("Mattermost-User-Id")))
}

// act applies an action for userID and tells everybody. A refused action (see
// isRefusal) is the user's doing, not a failure: the piggy is still returned.
func (p *Plugin) act(action, userID string) (pet, error) {
	schedule := p.settings.Load().schedule
	before, after, err := p.update(func(pt *pet, now time.Time) error { return pt.apply(action, userID, now, schedule) })
	if err != nil && !isRefusal(err) {
		return pet{}, err
	}
	p.changed(before, after, actionIfDone(action, err), userID)
	return after, err
}

func actionIfDone(action string, refusal error) string {
	if refusal != nil {
		return ""
	}
	return action
}

func isRefusal(err error) bool {
	for _, refusal := range []error{errDead, errAlive, errAsleep, errFull, errTired, errBadInput, errCooldown, errAwake} {
		if errors.Is(err, refusal) {
			return true
		}
	}
	return false
}

func (p *Plugin) handleAction(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Action string `json:"action"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024)).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	after, err := p.act(req.Action, r.Header.Get("Mattermost-User-Id"))
	if err != nil && !isRefusal(err) {
		p.API.LogError("Failed to update the piggy", "action", req.Action, "err", err.Error())
		http.Error(w, "could not update the piggy", http.StatusInternalServerError)
		return
	}
	resp := map[string]any{"pet": p.view(&after, time.Now(), r.Header.Get("Mattermost-User-Id"))}
	if err != nil {
		resp["refused"] = err.Error()
	}
	writeJSON(w, http.StatusOK, resp)
}
