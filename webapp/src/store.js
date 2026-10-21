// The piggy as the server last sent it, shared by the widget and its panel.
// Server state arrives from the API and from websocket events; reactions are
// the actions anybody just did, so every open widget plays them.

let pet = null;
const listeners = new Set();
const reactionListeners = new Set();

export function getPet() {
    return pet;
}

export function setPet(next) {
    pet = next;
    listeners.forEach((l) => l());
}

export function subscribe(listener) {
    listeners.add(listener);
    return () => listeners.delete(listener);
}

export function onReaction(listener) {
    reactionListeners.add(listener);
    return () => reactionListeners.delete(listener);
}

export function emitReaction(action) {
    reactionListeners.forEach((l) => l(action));
}

// handleStateEvent is the websocket handler of the server's "state" event.
// Broadcasts carry no cooldowns (they are per user), so the last known ones are kept.
export function handleStateEvent(msg) {
    setPet({...JSON.parse(msg.data.state), cooldowns: pet?.cooldowns});
    if (msg.data.action) {
        emitReaction(msg.data.action);
    }
}
