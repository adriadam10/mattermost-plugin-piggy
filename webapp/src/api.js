import manifest from '../../plugin.json';

// Mattermost rejects cookie-authenticated POSTs without its CSRF header.
function csrfToken() {
    return document.cookie.match(/(?:^|;\s*)MMCSRF=([^;]+)/)?.[1] ?? '';
}

async function request(method, path, body) {
    const response = await fetch(`${window.basename || ''}/plugins/${manifest.id}/api/v1/${path}`, {
        method,
        credentials: 'include',
        headers: {
            'Content-Type': 'application/json',
            'X-Requested-With': 'XMLHttpRequest',
            'X-CSRF-Token': csrfToken(),
        },
        body: body && JSON.stringify(body),
    });
    if (!response.ok) {
        throw new Error(`${manifest.id}: ${method} ${path} -> ${response.status}`);
    }
    return response.json();
}

export const fetchPet = () => request('GET', 'pet');
// Resolves to {pet, refused?}: refused is why the piggy did not want to (full, asleep...).
export const doAction = (action) => request('POST', 'action', {action});
// Saves the same preference as Settings → Piggy; Mattermost then updates every open window.
export const setShown = (show) => request('POST', 'visibility', {show});
