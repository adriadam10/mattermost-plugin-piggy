import React from 'react';

import {doAction, setShown} from './api.js';
import {placement} from './placement.js';
import {age, ago} from './i18n.js';
import {setPet} from './store.js';

const {useEffect, useLayoutEffect, useRef, useState} = React;

const ACTIONS = [
    {action: 'feed', icon: '🍎'},
    {action: 'pet', icon: '💕'},
    {action: 'play', icon: '⚽'},
    {action: 'sleep', icon: '💤'},
];
const STATS = ['fullness', 'happiness', 'energy', 'health'];
const SEGMENTS = 10;
const LOG_SHOWN = 4;
// Cooldowns are shown in whole minutes, so a refresh every few seconds is enough.
const CLOCK_MS = 5000;

// minutesLeft rounds up: "1m" until the button is really ready.
const minutesLeft = (until, now) => Math.ceil((until - now) / 60000);

function level(value) {
    if (value > 50) {
        return 'good';
    }
    return value > 25 ? 'low' : 'bad';
}

function Bar({label, value}) {
    const filled = Math.round(value / (100 / SEGMENTS));
    return (
        <div className='piggy-stat'>
            <span className='piggy-stat-label'>{label}</span>
            <span
                className={`piggy-bar piggy-bar-${level(value)}`}
                role='meter'
                aria-label={label}
                aria-valuemin={0}
                aria-valuemax={100}
                aria-valuenow={Math.round(value)}
            >
                {Array.from({length: SEGMENTS}, (_, i) => (
                    <span
                        key={i}
                        className={i < filled ? 'on' : ''}
                    />
                ))}
            </span>
        </div>
    );
}

export default function Panel({pet, error, t, locale, onClose, anchor}) {
    const ref = useRef(null);
    const [busy, setBusy] = useState(false);
    const [notice, setNotice] = useState('');
    const [now, setNow] = useState(Date.now());

    useEffect(() => {
        const timer = setInterval(() => setNow(Date.now()), CLOCK_MS);
        const onKey = (e) => e.key === 'Escape' && onClose();
        document.addEventListener('keydown', onKey);
        return () => {
            clearInterval(timer);
            document.removeEventListener('keydown', onKey);
        };
    }, [onClose]);

    // Placed once its size is known, and again whenever its content changes size.
    useLayoutEffect(() => {
        const el = ref.current;
        const place = placement(anchor, el.offsetWidth, el.offsetHeight, window.innerWidth, window.innerHeight);
        el.style.left = `${place.left}px`;
        el.style.top = `${place.top}px`;
    });

    const act = async (action) => {
        setBusy(true);
        setNotice('');
        try {
            const res = await doAction(action);
            setPet(res.pet);
            if (res.refused) {
                setNotice(t(`refused_${res.refused}`));
            }
        } catch {
            setNotice(t('loadError'));
        } finally {
            setBusy(false);
        }
    };

    const dead = pet?.mood === 'dead';
    return (
        <div
            ref={ref}
            className='piggy-panel'
            role='dialog'
            aria-label={pet?.name}
        >
            <div className='piggy-head'>
                <strong>{pet ? pet.name : '...'}</strong>
                {pet && <span className={`piggy-mood piggy-mood-${pet.mood}`}>{t(`mood_${pet.mood}`)}</span>}
                <button
                    className='piggy-close'
                    aria-label={t('close')}
                    onClick={onClose}
                >{'×'}</button>
            </div>
            {!pet && <p className='piggy-notice'>{error ? t('loadError') : '...'}</p>}
            {pet && dead && (
                <div className='piggy-rip'>
                    <div className='piggy-rip-title'>{t('rip')}</div>
                    <div>{`${t('diedAfter')} ${age(pet.born_at, pet.died_at, locale)}`}</div>
                </div>
            )}
            {pet && !dead && STATS.map((s) => (
                <Bar
                    key={s}
                    label={t(s)}
                    value={pet[s]}
                />
            ))}
            {pet && (
                <div className='piggy-actions'>
                    {dead ? (
                        <button
                            className='piggy-btn piggy-btn-wide'
                            disabled={busy}
                            onClick={() => act('adopt')}
                        >{`🐷 ${t('adopt')}`}</button>
                    ) : ACTIONS.map(({action, icon}) => {
                        const until = pet.cooldowns?.[action] ?? 0;
                        const waiting = until > now;
                        return (
                            <button
                                key={action}
                                className='piggy-btn'
                                disabled={busy || waiting}
                                title={waiting ? t('cooldown') : undefined}
                                onClick={() => act(action)}
                            >{waiting ? `${icon} ${t(action)} · ${minutesLeft(until, now)}m` : `${icon} ${t(action)}`}</button>
                        );
                    })}
                </div>
            )}
            {notice && <p className='piggy-notice'>{notice}</p>}
            {pet?.log.length > 0 && (
                <ul className='piggy-log'>
                    {pet.log.slice(0, LOG_SHOWN).map((e) => (
                        <li key={`${e.at}-${e.action}-${e.username}`}>
                            <span>{`@${e.username} ${t(`did_${e.action}`)}`}</span>
                            <time>{ago(e.at, locale)}</time>
                        </li>
                    ))}
                </ul>
            )}
            {pet && (
                <div className='piggy-foot'>
                    <span>{`${t('bestFriend')}: ${pet.best_friend ? `@${pet.best_friend}` : t('nobody')}`}</span>
                    {!dead && <span>{`${t('age')}: ${age(pet.born_at, Date.now(), locale)}`}</span>}
                    <span>{t('tip', {trigger: pet.trigger})}</span>
                    <button
                        className='piggy-hide'
                        onClick={() => setShown(false).catch(() => setNotice(t('loadError')))}
                    >{t('hideMe', {trigger: pet.trigger})}</button>
                </div>
            )}
        </div>
    );
}
