package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const (
	petKey    = "pet"
	kvRetries = 5
)

func (p *Plugin) loadPet() (pet, []byte, error) {
	raw, appErr := p.API.KVGet(petKey)
	if appErr != nil {
		return pet{}, nil, appErr
	}
	if raw == nil {
		return newPet(1, time.Now()), nil, nil
	}
	var pt pet
	if err := json.Unmarshal(raw, &pt); err != nil {
		return pet{}, nil, fmt.Errorf("decode pet: %w", err)
	}
	return pt, raw, nil
}

// update brings the piggy up to date, applies change and saves it with a
// compare-and-set loop, since several people (and the ticker) can act at once.
// It returns the piggy before and after, and change's error, if any.
func (p *Plugin) update(change func(*pet, time.Time) error) (before, after pet, err error) {
	s := p.settings.Load()
	for range kvRetries {
		now := time.Now()
		pt, old, err := p.loadPet()
		if err != nil {
			return pet{}, pet{}, err
		}
		before = pt
		pt.advance(now, s.schedule)
		changeErr := change(&pt, now)
		if changeErr != nil {
			// A refused action is not saved, but the time that went by still is.
			pt = before
			pt.advance(now, s.schedule)
		}
		updated, err := json.Marshal(pt)
		if err != nil {
			return pet{}, pet{}, err
		}
		ok, appErr := p.API.KVCompareAndSet(petKey, old, updated)
		if appErr != nil {
			return pet{}, pet{}, appErr
		}
		if ok {
			return before, pt, changeErr
		}
	}
	return pet{}, pet{}, errors.New("the piggy kept changing while saving it")
}
