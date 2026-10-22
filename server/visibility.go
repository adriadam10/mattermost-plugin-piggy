package main

import (
	"encoding/json"
	"net/http"

	"github.com/mattermost/mattermost/server/public/model"
)

const (
	manifestID = "com.github.adriadam10.piggy"
	// The webapp registers this user setting (Settings → Piggy); hiding from the
	// panel or the slash command writes the same preference.
	showSetting = "show_piggy"
	showValue   = "show"
	hideValue   = "hide"
	// Length of the category column of the preferences table; the webapp cuts to it too.
	maxCategoryLength = 32
)

// prefCategory is where Mattermost keeps the plugin's user settings: "pp_" plus
// the plugin id, cut to the 32 characters the preferences table allows.
func prefCategory() string {
	category := "pp_" + manifestID
	return category[:min(len(category), maxCategoryLength)]
}

// setShown saves whether userID sees the piggy. Mattermost tells the user's
// clients, so open windows show or hide it at once.
func (p *Plugin) setShown(userID string, shown bool) error {
	value := hideValue
	if shown {
		value = showValue
	}
	if appErr := p.API.UpdatePreferencesForUser(userID, []model.Preference{{
		UserId:   userID,
		Category: prefCategory(),
		Name:     showSetting,
		Value:    value,
	}}); appErr != nil {
		return appErr
	}
	return nil
}

func (p *Plugin) handleVisibility(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Show bool `json:"show"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024)).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := p.setShown(r.Header.Get("Mattermost-User-Id"), req.Show); err != nil {
		p.API.LogError("Failed to save the piggy visibility", "err", err.Error())
		http.Error(w, "could not save", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"show": req.Show})
}
