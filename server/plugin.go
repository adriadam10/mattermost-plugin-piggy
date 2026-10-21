package main

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
	// The distroless Mattermost image has no zoneinfo.
	_ "time/tzdata"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

const (
	botUsername = "piggy"
	// Sent to every client when the piggy changes, so all widgets stay in step.
	stateEvent = "state"
	// The needs drop slowly, but naps end and the night starts at a given minute:
	// clients learn it from this tick.
	tickInterval = time.Minute
)

type Plugin struct {
	plugin.MattermostPlugin
	router   http.Handler
	settings atomic.Pointer[settings]
	botID    string
	// Resolved from the AnnouncementChannel setting; empty when announcements are off.
	announceChannelID atomic.Value
	stop              chan struct{}
	ticker            sync.WaitGroup
}

func (p *Plugin) OnActivate() error {
	s := p.settings.Load()
	if s == nil {
		return errors.New("no valid configuration")
	}
	botID, err := p.API.EnsureBotUser(&model.Bot{Username: botUsername, DisplayName: s.name, Description: "The team piggy"})
	if err != nil {
		return fmt.Errorf("ensure bot: %w", err)
	}
	p.botID = botID
	if avatar, err := avatarPNG(); err != nil {
		p.API.LogWarn("Failed to draw the bot avatar", "err", err.Error())
	} else if appErr := p.API.SetProfileImage(botID, avatar); appErr != nil {
		p.API.LogWarn("Failed to set the bot avatar", "err", appErr.Error())
	}
	if err := p.applySettings(nil, s); err != nil {
		return err
	}
	p.router = p.newRouter()
	p.stop = make(chan struct{})
	p.ticker.Add(1)
	go p.tick()
	return nil
}

func (p *Plugin) OnDeactivate() error {
	close(p.stop)
	p.ticker.Wait()
	return nil
}

// applySettings registers the slash command and resolves the announcement
// channel; old is nil on activation.
func (p *Plugin) applySettings(old, s *settings) error {
	if old != nil && old.trigger != s.trigger {
		if err := p.API.UnregisterCommand("", old.trigger); err != nil {
			p.API.LogWarn("Failed to unregister the old slash command", "trigger", old.trigger, "err", err.Error())
		}
	}
	t := translatorFor(*p.API.GetConfig().LocalizationSettings.DefaultClientLocale)
	if err := p.API.RegisterCommand(&model.Command{
		Trigger:          s.trigger,
		DisplayName:      "Piggy",
		Description:      t("desc"),
		AutoComplete:     true,
		AutoCompleteDesc: t("desc"),
		AutoCompleteHint: t("hint"),
	}); err != nil {
		return fmt.Errorf("register /%s: %w", s.trigger, err)
	}
	if old != nil && old.name != s.name {
		if _, err := p.API.EnsureBotUser(&model.Bot{Username: botUsername, DisplayName: s.name}); err != nil {
			p.API.LogWarn("Failed to rename the bot", "err", err.Error())
		}
	}
	p.announceChannelID.Store(p.resolveAnnounceChannel(s))
	return nil
}

// resolveAnnounceChannel finds the channel and puts the bot in it. A wrong
// channel only turns announcements off: the piggy itself keeps working.
func (p *Plugin) resolveAnnounceChannel(s *settings) string {
	if s.announceTeam == "" {
		return ""
	}
	team, appErr := p.API.GetTeamByName(s.announceTeam)
	if appErr != nil {
		p.API.LogError("Announcement team not found", "team", s.announceTeam, "err", appErr.Error())
		return ""
	}
	channel, appErr := p.API.GetChannelByName(team.Id, s.announceChannel, false)
	if appErr != nil {
		p.API.LogError("Announcement channel not found", "channel", s.announceChannel, "err", appErr.Error())
		return ""
	}
	if _, appErr := p.API.CreateTeamMember(team.Id, p.botID); appErr != nil {
		p.API.LogWarn("Failed to add the bot to the announcement team", "err", appErr.Error())
	}
	if _, appErr := p.API.AddChannelMember(channel.Id, p.botID); appErr != nil {
		p.API.LogWarn("Failed to add the bot to the announcement channel", "err", appErr.Error())
	}
	return channel.Id
}

func (p *Plugin) ServeHTTP(_ *plugin.Context, w http.ResponseWriter, r *http.Request) {
	p.router.ServeHTTP(w, r)
}

func (p *Plugin) tick() {
	defer p.ticker.Done()
	t := time.NewTicker(tickInterval)
	defer t.Stop()
	for {
		select {
		case <-p.stop:
			return
		case <-t.C:
			before, after, err := p.update(func(*pet, time.Time) error { return nil })
			if err != nil {
				p.API.LogError("Failed to update the piggy", "err", err.Error())
				continue
			}
			p.changed(before, after, "", "")
		}
	}
}

// changed tells every client and, on a turn for the worse, the announcement channel.
func (p *Plugin) changed(before, after pet, action, userID string) {
	p.announce(&before, &after)
	view := p.view(&after, time.Now(), "")
	p.API.PublishWebSocketEvent(stateEvent, map[string]any{
		"state":   mustJSON(view),
		"action":  action,
		"user_id": userID,
	}, &model.WebsocketBroadcast{})
}

func (p *Plugin) announce(before, after *pet) {
	channelID, _ := p.announceChannelID.Load().(string)
	if channelID == "" {
		return
	}
	s := p.settings.Load()
	t := translatorFor(*p.API.GetConfig().LocalizationSettings.DefaultServerLocale)
	name := petName(s.name, after.Generation)
	switch {
	case !before.dead() && after.dead() && before.Generation == after.Generation:
		p.say(channelID, t("died", name, age(t, after.BornAt, after.DiedAt), s.trigger))
	case after.Health < sickHealth && !after.SickAnnounced && !after.dead():
		p.say(channelID, t("sick", name))
		p.markSickAnnounced(true)
	case after.Health > recoveredHealth && after.SickAnnounced:
		p.markSickAnnounced(false)
	}
}

func (p *Plugin) markSickAnnounced(announced bool) {
	if _, _, err := p.update(func(pt *pet, _ time.Time) error {
		pt.SickAnnounced = announced
		return nil
	}); err != nil {
		p.API.LogError("Failed to save the sick announcement", "err", err.Error())
	}
}

func (p *Plugin) say(channelID, message string) bool {
	if _, appErr := p.API.CreatePost(&model.Post{UserId: p.botID, ChannelId: channelID, Message: message}); appErr != nil {
		p.API.LogError("Failed to post as the piggy", "channel_id", channelID, "err", appErr.Error())
		return false
	}
	return true
}

func main() {
	plugin.ClientMain(&Plugin{})
}
