import assert from 'node:assert/strict';
import {test} from 'node:test';

import {age, ago, keysOf, languages, translator} from './i18n.js';

test('every language has every string', () => {
    for (const lang of languages) {
        assert.deepEqual(keysOf(lang).sort(), keysOf('en').sort(), lang);
    }
});

test('falls back to English and fills placeholders', () => {
    const t = translator('fr-FR');
    assert.equal(t('feed'), 'Feed');
    assert.equal(t('tip', {trigger: 'piggy'}), 'Also /piggy in any channel.');
    assert.equal(translator('es')('feed'), 'Comer');
});

test('relative times and age', () => {
    const now = Date.UTC(2026, 9, 21, 12);
    assert.equal(ago(now - 5 * 60000, 'en', now), '5 minutes ago');
    assert.equal(ago(now - 20000, 'en', now), 'this minute');
    assert.equal(age(now - 3 * 86400000, now, 'en'), '3 days');
    assert.equal(age(now - 5 * 3600000, now, 'es'), '5 horas');
});
