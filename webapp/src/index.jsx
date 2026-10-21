import React from 'react';
import {useSelector} from 'react-redux';

import manifest from '../../plugin.json';

import {fetchPet} from './api.js';
import {translator} from './i18n.js';
import {emitReaction, handleStateEvent, setPet} from './store.js';
import {STYLES} from './styles.js';
import Widget from './widget.jsx';

// The user setting registered below; Mattermost stores it as a preference.
const SHOW_SETTING = 'show_piggy';

function currentLocale(state) {
    const {currentUserId, profiles} = state.entities.users;
    return profiles[currentUserId]?.locale || document.documentElement.lang || 'en';
}

class Plugin {
    initialize(registry, store) {
        const locale = currentLocale(store.getState());
        const t = translator(locale);

        this.style = document.createElement('style');
        this.style.textContent = STYLES;
        document.head.appendChild(this.style);

        registry.registerRootComponent(() => {
            const shown = useSelector((state) => state.entities.preferences.myPreferences[`pp_${manifest.id}--${SHOW_SETTING}`]?.value) !== 'hide';
            return shown ? (
                <Widget
                    t={t}
                    locale={locale}
                />
            ) : null;
        });
        registry.registerWebSocketEventHandler(`custom_${manifest.id}_state`, handleStateEvent);
        // Somebody else writes in the channel on screen: the piggy pricks up its ears.
        registry.registerWebSocketEventHandler('posted', (msg) => {
            const state = store.getState();
            const post = JSON.parse(msg.data.post);
            if (post.channel_id === state.entities.channels.currentChannelId && post.user_id !== state.entities.users.currentUserId) {
                emitReaction('notice');
            }
        });
        // Events missed while offline are gone; the server has the current piggy.
        registry.registerReconnectHandler(() => fetchPet().then(setPet, (err) => console.warn(err))); // eslint-disable-line no-console
        registry.registerUserSettings({
            id: manifest.id,
            uiName: t('settingsSection'),
            sections: [{
                title: t('settingsSection'),
                settings: [{
                    name: SHOW_SETTING,
                    type: 'radio',
                    title: t('settingTitle'),
                    helpText: t('settingHelp'),
                    default: 'show',
                    options: [{value: 'show', text: t('show')}, {value: 'hide', text: t('hide')}],
                }],
            }],
        });
    }

    uninitialize() {
        this.style?.remove();
    }
}

window.registerPlugin(manifest.id, new Plugin());
