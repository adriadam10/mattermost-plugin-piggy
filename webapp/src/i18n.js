const STRINGS = {
    en: {
        feed: 'Feed',
        pet: 'Pet',
        play: 'Play',
        sleep: 'Sleep',
        adopt: 'Adopt a new piggy',
        fullness: 'Food',
        happiness: 'Joy',
        energy: 'Energy',
        health: 'Health',
        close: 'Close',
        bestFriend: 'Best friend',
        nobody: 'nobody yet',
        age: 'Age',
        diedAfter: 'Died after',
        rip: 'R.I.P.',
        loadError: 'The piggy could not be reached.',
        tip: 'Also /{trigger} in any channel.',
        hideMe: 'Hide the piggy for me (/{trigger} show brings it back)',
        settingsSection: 'Piggy',
        settingTitle: 'Show the piggy',
        settingHelp: 'The team piggy walks round the edges of the window. Hiding it only affects you; the link at the bottom of its panel does the same.',
        show: 'Show',
        hide: 'Hide',
        did_feed: 'fed it',
        did_pet: 'petted it',
        did_play: 'played with it',
        did_sleep: 'put it to bed',
        did_adopt: 'adopted it',
        refused_full: 'It is full.',
        refused_asleep: 'Shh... it is sleeping.',
        refused_tired: 'Too tired to play. Let it sleep.',
        refused_dead: 'It is dead.',
        refused_alive: 'It is still alive.',
        refused_unknown: 'It did not understand that.',
        refused_cooldown: 'You just did that. Give it a break.',
        refused_not_sleepy: 'It is not sleepy. It naps on its own when tired.',
        cooldown: 'You can do this again in a few minutes; somebody else can do it now.',
        mood_dead: 'dead',
        mood_sleeping: 'sleeping',
        mood_night: 'off work, napping',
        mood_sick: 'sick!',
        mood_hungry: 'hungry',
        mood_sad: 'sad',
        mood_tired: 'tired',
        mood_happy: 'happy',
        mood_ok: 'fine',
    },
    es: {
        feed: 'Comer',
        pet: 'Mimar',
        play: 'Jugar',
        sleep: 'Dormir',
        adopt: 'Adoptar otro cerdito',
        fullness: 'Comida',
        happiness: 'Alegría',
        energy: 'Energía',
        health: 'Salud',
        close: 'Cerrar',
        bestFriend: 'Mejor amigo',
        nobody: 'nadie todavía',
        age: 'Edad',
        diedAfter: 'Murió tras',
        rip: 'D.E.P.',
        loadError: 'No se ha podido contactar con el cerdito.',
        tip: 'También con /{trigger} en cualquier canal.',
        hideMe: 'Ocultar el cerdito para mí (/{trigger} show lo vuelve a traer)',
        settingsSection: 'Cerdito',
        settingTitle: 'Mostrar el cerdito',
        settingHelp: 'El cerdito del equipo pasea por los bordes de la ventana. Ocultarlo solo te afecta a ti; el enlace al pie de su panel hace lo mismo.',
        show: 'Mostrar',
        hide: 'Ocultar',
        did_feed: 'le dio de comer',
        did_pet: 'le hizo mimos',
        did_play: 'jugó con él',
        did_sleep: 'lo mandó a dormir',
        did_adopt: 'lo adoptó',
        refused_full: 'Está lleno.',
        refused_asleep: 'Chsss... está durmiendo.',
        refused_tired: 'Está muy cansado para jugar. Déjalo dormir.',
        refused_dead: 'Ha muerto.',
        refused_alive: 'Sigue vivo.',
        refused_unknown: 'No lo ha entendido.',
        refused_cooldown: 'Eso ya lo acabas de hacer. Déjalo respirar.',
        refused_not_sleepy: 'No tiene sueño. Ya se echará la siesta él solo.',
        cooldown: 'Podrás repetirlo en unos minutos; otra persona puede hacerlo ya.',
        mood_dead: 'muerto',
        mood_sleeping: 'durmiendo',
        mood_night: 'fuera de horario, sesteando',
        mood_sick: '¡enfermo!',
        mood_hungry: 'hambriento',
        mood_sad: 'triste',
        mood_tired: 'cansado',
        mood_happy: 'feliz',
        mood_ok: 'bien',
    },
};

export function translator(locale) {
    const strings = STRINGS[(locale || '').split('-')[0]] || STRINGS.en;
    return (key, vars = {}) => (strings[key] ?? key).replace(/\{(\w+)\}/g, (_, name) => vars[name] ?? '');
}

export const languages = Object.keys(STRINGS);
export const keysOf = (lang) => Object.keys(STRINGS[lang]);

const UNITS = [['day', 86400], ['hour', 3600], ['minute', 60]];

// ago is "5 minutes ago", "yesterday"... in the user's language.
export function ago(ms, locale, now = Date.now()) {
    const seconds = (ms - now) / 1000;
    const format = new Intl.RelativeTimeFormat(locale, {numeric: 'auto'});
    for (const [unit, size] of UNITS) {
        if (Math.abs(seconds) >= size) {
            return format.format(Math.round(seconds / size), unit);
        }
    }
    return format.format(0, 'minute');
}

// age is how long the piggy has lived, in days or hours.
export function age(fromMs, toMs, locale) {
    const hours = Math.max(0, (toMs - fromMs) / 3600000);
    const format = new Intl.NumberFormat(locale, {style: 'unit', unitDisplay: 'long', unit: hours >= 24 ? 'day' : 'hour', maximumFractionDigits: 0});
    return format.format(hours >= 24 ? Math.floor(hours / 24) : Math.floor(hours));
}
