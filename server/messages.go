package main

import (
	"fmt"
	"strings"
)

// Texts the server writes: slash command replies and channel announcements.
var messages = map[string]map[string]string{
	"en": {
		"fed":            "%s gives %s an apple. *Nom nom!* 🍎",
		"petted":         "%s scratches %s behind the ears. *Oink!* 💕",
		"played":         "%s plays ball with %s. *Wheee!* ⚽",
		"slept":          "%s tucks %s in for a nap. 💤",
		"adopted":        "%s adopts %s. Welcome! 🐷",
		"err_dead":       "%s is dead. 🪦 Adopt a new piggy with `/%s adopt`.",
		"err_alive":      "%s is alive and well, no need to adopt another one.",
		"err_asleep":     "Shh... %s is sleeping. 💤",
		"err_full":       "%s is full and cannot eat another bite.",
		"err_tired":      "%s is too tired to play. Let it sleep a bit.",
		"err_unknown":    "Unknown action. Try `/%s help`.",
		"err_cooldown":   "You just did that. Give %s a break and try again in %d min.",
		"hiden":          "Done, %s will not walk on your screen any more (only for you). `/%s show` brings it back.",
		"shown":          "%s is back on your screen. 🐷 `/%s hide` to hide it.",
		"err_not_sleepy": "%s is not sleepy at all. It will nap on its own when it is tired.",
		"sick":           "🤒 %s is getting sick from neglect! Feed it and play with it before it is too late.",
		"died":           "🪦 %s has died after %s. Rest in peace, little piggy. Adopt a new one with `/%s adopt`.",
		"days":           "%d days",
		"one_day":        "1 day",
		"less_day":       "less than a day",
		"status":         "**%s** · %s\n| | |\n|:--|:--|\n| 🍎 Fullness | %s |\n| 💕 Happiness | %s |\n| ⚡ Energy | %s |\n| ❤️ Health | %s |\n\nAge: %s · Best friend: %s",
		"status_dead":    "🪦 **%s** died after %s. Best friend: %s. Adopt a new piggy with `/%s adopt`.",
		"nobody":         "nobody yet",
		"help":           "`/%[1]s` shows how the piggy is doing. Also `/%[1]s feed`, `/%[1]s pet`, `/%[1]s play`, `/%[1]s sleep` and, if it died, `/%[1]s adopt`. Click on the piggy for the same. `/%[1]s hide` takes it off your screen and `/%[1]s show` brings it back.",
		"desc":           "Look after the team piggy",
		"hint":           "[status|feed|pet|play|sleep|adopt|hide|show|help]",
		"mood_dead":      "dead",
		"mood_sleeping":  "sleeping",
		"mood_night":     "off work, napping",
		"mood_sick":      "sick",
		"mood_hungry":    "hungry",
		"mood_sad":       "sad",
		"mood_tired":     "tired",
		"mood_happy":     "happy",
		"mood_ok":        "fine",
	},
	"es": {
		"fed":            "%s le da una manzana a %s. *¡Ñam ñam!* 🍎",
		"petted":         "%s le rasca detrás de las orejas a %s. *¡Oink!* 💕",
		"played":         "%s juega a la pelota con %s. *¡Yupiii!* ⚽",
		"slept":          "%s arropa a %s para que eche una siesta. 💤",
		"adopted":        "%s adopta a %s. ¡Bienvenido! 🐷",
		"err_dead":       "%s ha muerto. 🪦 Adopta otro cerdito con `/%s adopt`.",
		"err_alive":      "%s está vivo y coleando, no hace falta adoptar otro.",
		"err_asleep":     "Chsss... %s está durmiendo. 💤",
		"err_full":       "%s está lleno y no le cabe ni un bocado más.",
		"err_tired":      "%s está demasiado cansado para jugar. Déjalo dormir un rato.",
		"err_unknown":    "Acción desconocida. Prueba `/%s help`.",
		"err_cooldown":   "Eso ya lo acabas de hacer. Deja respirar a %s y vuelve a probar en %d min.",
		"hiden":          "Hecho, %s ya no pasea por tu pantalla (solo para ti). `/%s show` lo vuelve a traer.",
		"shown":          "%s vuelve a tu pantalla. 🐷 `/%s hide` para ocultarlo.",
		"err_not_sleepy": "%s no tiene nada de sueño. Ya se echará la siesta él solo cuando esté cansado.",
		"sick":           "🤒 ¡%s se está poniendo malo de tanto abandono! Dadle de comer y jugad con él antes de que sea tarde.",
		"died":           "🪦 %s ha muerto tras %s de vida. Descansa en paz, cerdito. Adoptad otro con `/%s adopt`.",
		"days":           "%d días",
		"one_day":        "1 día",
		"less_day":       "menos de un día",
		"status":         "**%s** · %s\n| | |\n|:--|:--|\n| 🍎 Saciedad | %s |\n| 💕 Felicidad | %s |\n| ⚡ Energía | %s |\n| ❤️ Salud | %s |\n\nEdad: %s · Mejor amigo: %s",
		"status_dead":    "🪦 **%s** murió tras %s de vida. Mejor amigo: %s. Adopta otro cerdito con `/%s adopt`.",
		"nobody":         "nadie todavía",
		"help":           "`/%[1]s` enseña cómo está el cerdito. También `/%[1]s feed` (comer), `/%[1]s pet` (acariciar), `/%[1]s play` (jugar), `/%[1]s sleep` (dormir) y, si ha muerto, `/%[1]s adopt` (adoptar). Lo mismo haciendo clic en el cerdito. `/%[1]s hide` lo quita de tu pantalla y `/%[1]s show` lo vuelve a traer.",
		"desc":           "Cuida del cerdito del equipo",
		"hint":           "[status|feed|pet|play|sleep|adopt|hide|show|help]",
		"mood_dead":      "muerto",
		"mood_sleeping":  "durmiendo",
		"mood_night":     "fuera de horario, sesteando",
		"mood_sick":      "enfermo",
		"mood_hungry":    "hambriento",
		"mood_sad":       "triste",
		"mood_tired":     "cansado",
		"mood_happy":     "feliz",
		"mood_ok":        "bien",
	},
}

type translator func(key string, args ...any) string

func translatorFor(locale string) translator {
	strs, ok := messages[strings.SplitN(locale, "-", 2)[0]]
	if !ok {
		strs = messages["en"]
	}
	return func(key string, args ...any) string {
		return fmt.Sprintf(strs[key], args...)
	}
}

var romans = []struct {
	value  int
	symbol string
}{{1000, "M"}, {900, "CM"}, {500, "D"}, {400, "CD"}, {100, "C"}, {90, "XC"}, {50, "L"}, {40, "XL"}, {10, "X"}, {9, "IX"}, {5, "V"}, {4, "IV"}, {1, "I"}}

func roman(n int) string {
	var b strings.Builder
	for _, r := range romans {
		for n >= r.value {
			b.WriteString(r.symbol)
			n -= r.value
		}
	}
	return b.String()
}

// petName is the configured name plus the generation from the second piggy on: Bacon, Bacon II...
func petName(name string, generation int) string {
	if generation <= 1 {
		return name
	}
	return name + " " + roman(generation)
}

// bar draws a stat as ten blocks, the 8-bit way.
func bar(value float64) string {
	filled := int(value/10 + 0.5)
	return strings.Repeat("▰", filled) + strings.Repeat("▱", 10-filled) + fmt.Sprintf(" %d", int(value+0.5))
}

func age(t translator, fromMs, toMs int64) string {
	days := int((toMs - fromMs) / (24 * 60 * 60 * 1000))
	switch days {
	case 0:
		return t("less_day")
	case 1:
		return t("one_day")
	default:
		return t("days", days)
	}
}
