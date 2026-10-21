# Mattermost Piggy

An 8-bit team pet for Mattermost: one piggy, shared by the whole server, that walks along the bottom
left corner of the webapp. Everybody sees the same piggy, and if nobody looks after it, it dies.

- **Click the piggy** to open its panel: food, joy, energy and health bars, and buttons to feed it,
  pet it, play with it or send it to bed. Drag it to move it (the position is remembered per browser).
- **Everybody sees every action live**: when someone feeds it, the piggy eats on every open screen.
- **`/piggy`** in any channel shows its status; `/piggy feed|pet|play|sleep` does the same as the
  buttons and the piggy's bot says it in the channel. `/piggy adopt` brings a new one after a death.
- **It only gets hungry during working hours** (Monday to Friday, configurable hours and time zone).
  Nights and weekends do not count, and it sleeps then.
- **It can die**: after *Working days to death* (default 3) without anybody feeding or playing with
  it. It gets hungry after a third of that, then sick, and an optional announcement channel is told
  when it gets sick and when it dies. The next piggy is *Bacon II*, *Bacon III*...
- Each user can hide it in **Settings → Piggy**. It is hidden on phone-sized screens; the slash
  command still works there and in the mobile apps.
- English and Spanish, following the user's language.

## How it works

The server part keeps the piggy in the plugin's KV store and brings its needs up to date every five
minutes and on every request, counting only working time. Changes go to every client over the
websocket. The webapp draws the piggy on a small canvas scaled up without smoothing; walking,
sleeping and dying are pixel edits of the sprite sheet's frames.

## Settings

| Setting | Default | |
|---|---|---|
| Name | `Bacon` | Generations after a death get a Roman numeral. |
| Slash command | `piggy` | |
| Working days to death | `3` | 1 to 30. |
| Workday start / end | `8` / `18` | Hours in the time zone below. |
| Time zone | `UTC` | IANA name, e.g. `Europe/Madrid`. |
| Announcement channel | empty | `team-name/channel-name`; the bot joins it. |

## Install

Requires Mattermost 11.11 or later. Download the `.tar.gz` of the latest release and upload it in
**System Console → Plugins → Plugin Management**, or:

```sh
mmctl plugin install-url https://github.com/adriadam10/mattermost-plugin-piggy/releases/download/v0.2.1/com.github.adriadam10.piggy-0.2.1.tar.gz
mmctl plugin enable com.github.adriadam10.piggy
```

## Build

`./build.sh` tests and builds everything in Go and Node containers (only Docker is needed) and
writes `dist/com.github.adriadam10.piggy-<version>.tar.gz`. Pushing a `vX.Y.Z` tag matching
`version` in `plugin.json` publishes a release.

## Credits

Piggy sprite sheet: [Pig_Look Sprite Sheet 32x32](https://opengameart.org/content/piglook-sprite-sheet-32x32)
by Natural_Privateer, CC0. Recoloured in pink at run time.

## License

MIT
