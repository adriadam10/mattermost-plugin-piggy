package main

import (
	"errors"
	"strings"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

// What the bot says in the channel after each action.
var actionMessages = map[string]string{
	actionFeed:  "fed",
	actionPet:   "petted",
	actionPlay:  "played",
	actionSleep: "slept",
	actionAdopt: "adopted",
}

func ephemeral(text string) *model.CommandResponse {
	return &model.CommandResponse{ResponseType: model.CommandResponseTypeEphemeral, Text: text}
}

func (p *Plugin) ExecuteCommand(_ *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	s := p.settings.Load()
	locale := ""
	if user, appErr := p.API.GetUser(args.UserId); appErr == nil {
		locale = user.Locale
	}
	t := translatorFor(locale)

	fields := strings.Fields(args.Command)
	sub := "status"
	if len(fields) > 1 {
		sub = strings.ToLower(fields[1])
	}
	switch sub {
	case "help":
		return ephemeral(t("help", s.trigger)), nil
	case "status":
		_, after, err := p.update(func(*pet, time.Time) error { return nil })
		if err != nil {
			return nil, model.NewAppError("ExecuteCommand", "piggy.status", nil, err.Error(), 500)
		}
		return ephemeral(p.statusText(t, &after)), nil
	}
	if _, known := actionMessages[sub]; !known {
		return ephemeral(t("err_unknown", s.trigger)), nil
	}

	after, err := p.act(sub, args.UserId)
	name := petName(s.name, after.Generation)
	switch {
	case errors.Is(err, errDead):
		return ephemeral(t("err_dead", name, s.trigger)), nil
	case errors.Is(err, errCooldown):
		wait := time.Until(time.UnixMilli(after.cooldowns(args.UserId, time.Now())[sub]))
		return ephemeral(t("err_cooldown", name, int(wait.Minutes())+1)), nil
	case err != nil && isRefusal(err):
		return ephemeral(t("err_"+err.Error(), name)), nil
	case err != nil:
		return nil, model.NewAppError("ExecuteCommand", "piggy.action", nil, err.Error(), 500)
	}
	message := t(actionMessages[sub], "@"+p.username(args.UserId), name)
	// The bot may not be allowed in this channel (a DM between others); then only the user sees it.
	if !p.say(args.ChannelId, message) {
		return ephemeral(message), nil
	}
	return &model.CommandResponse{}, nil
}

func (p *Plugin) statusText(t translator, pt *pet) string {
	s := p.settings.Load()
	name := petName(s.name, pt.Generation)
	friend := t("nobody")
	if f := p.username(pt.bestFriend()); f != "" {
		friend = "@" + f
	}
	if pt.dead() {
		return t("status_dead", name, age(t, pt.BornAt, pt.DiedAt), friend, s.trigger)
	}
	return t("status", name, t("mood_"+pt.mood(time.Now(), s.schedule)),
		bar(pt.Fullness), bar(pt.Happiness), bar(pt.Energy), bar(pt.Health),
		age(t, pt.BornAt, time.Now().UnixMilli()), friend)
}
