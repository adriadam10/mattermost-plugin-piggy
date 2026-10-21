import {PALETTES, SIZE, overlay} from './frames.js';

// The canvas is a VIEW x VIEW square of game pixels with the piggy standing at
// the bottom and room above it for hearts and bubbles; CSS scales it up without
// smoothing. The square walks round the window edges (see track.js).
export const VIEW = 56;
export const TICK_MS = 125;
const PIG_X = (VIEW - SIZE) / 2;
// y of the piggy's frame: its last 4 rows are empty, so the hooves touch the edge.
const GROUND = VIEW - SIZE + 4;
// The piggy's rectangle in game pixels of the canvas, for clicks.
export const PIG_BOX = {x: PIG_X + 2, y: GROUND + 6, width: SIZE - 4, height: SIZE - 6};
const REACTION_TICKS = 18;
// Close enough to the mouse to sniff it, in CSS pixels along the edge.
const SNIFF_DISTANCE = 6;

const HEART = overlay('heart');
const ZZZ = overlay('zzz');
const APPLE = overlay('apple');
const BALL = overlay('ball');
const TEAR = overlay('tear');
const HALO = overlay('halo');
const BANG = overlay('bang');

const PINK = PALETTES.pink;
// Moods in which it gets up and wanders about.
const ROAMING = new Set(['ok', 'happy', 'tired', 'hungry']);

const lively = (mood) => mood === 'ok' || mood === 'happy';
const randomTicks = (min, max) => min + Math.floor(Math.random() * (max - min));

// Scene holds where the piggy is on the track and what it is doing; draw() runs every tick.
export default class Scene {
    constructor(frames, scale) {
        this.frames = frames;
        this.scale = scale;
        this.track = null;
        this.pos = 12;
        this.dir = 1;
        this.tick = 0;
        this.activity = {name: 'idle', until: 0};
        this.reaction = null;
        // Set by the widget: the panel is open, or the mouse waits near an edge.
        this.paused = false;
        this.target = null;
    }

    // react plays an action for a couple of seconds, whoever did it.
    react(action) {
        // A new message does not cut short somebody's action.
        if (action === 'notice' && this.reaction && this.reaction.action !== 'notice') {
            return;
        }
        this.reaction = {action, until: this.tick + REACTION_TICKS, start: this.tick};
    }

    pickActivity(mood) {
        const roll = Math.random();
        // A trip to anywhere on the edges, so it does not stay in one corner.
        if (this.track && roll < (lively(mood) ? 0.45 : 0.15)) {
            this.activity = {name: 'trip', to: Math.random() * this.track.length, until: this.tick + 2400};
        } else if (mood === 'happy' && roll < 0.6) {
            this.dir = Math.random() < 0.5 ? -1 : 1;
            this.activity = {name: 'run', until: this.tick + randomTicks(12, 40)};
        } else if (mood === 'happy' && roll < 0.7) {
            this.activity = {name: 'hop', until: this.tick + 12};
        } else if (roll < (lively(mood) ? 0.85 : 0.55)) {
            this.dir = Math.random() < 0.8 ? this.dir : -this.dir;
            this.activity = {name: 'walk', until: this.tick + randomTicks(20, 120)};
        } else if (roll < 0.85) {
            this.activity = {name: 'look', until: this.tick + 18};
        } else {
            this.activity = {name: 'idle', until: this.tick + randomTicks(8, 30)};
        }
    }

    // walk moves `pixels` game pixels along the track, unless the panel is open.
    walk(pixels) {
        if (!this.paused && this.track) {
            this.pos = this.track.wrap(this.pos + this.dir * pixels * this.scale);
        }
    }

    // front is the x of something `offset` pixels ahead of the snout (`width` wide).
    front(offset, width = 7) {
        return this.dir > 0 ? PIG_X + SIZE - 4 + offset : PIG_X + 4 - offset - width;
    }

    // chase walks towards the mouse; it returns a pose once there, or null.
    chase(mood) {
        const d = this.track.towards(this.pos, this.target);
        if (Math.abs(d) <= SNIFF_DISTANCE) {
            const k = this.tick;
            return {frame: k % 6 < 3 ? this.frames.sniff : this.frames.stand, palette: PINK, dy: 0, extras: k % 24 < 12 ? [[HEART, this.front(0), GROUND + 2 - ((k % 12) >> 1)]] : []};
        }
        // It trots to the mouse, so it gets there before losing interest.
        this.dir = Math.sign(d);
        this.walk(mood === 'tired' || mood === 'hungry' ? 2 : 5);
        return null;
    }

    reactionPose(mood) {
        const f = this.frames;
        const r = this.reaction;
        const k = this.tick - r.start;
        switch (r.action) {
        case 'feed':
            return {
                frame: k % 4 < 2 ? f.sniff : f.stand,
                palette: PINK,
                dy: 0,
                // The apple shrinks bite by bite in front of the snout.
                extras: k < REACTION_TICKS - 4 ? [[APPLE, this.front(2), GROUND + 21, Math.min(5, k >> 2)]] : [[HEART, this.front(-4), GROUND, 0]],
            };
        case 'pet':
            return {frame: k % 8 < 6 ? f.blink : f.stand, palette: PINK, dy: 0, extras: [[HEART, PIG_X + 12 + (k % 6 < 3 ? 0 : 1), GROUND + 4 - (k >> 1)]]};
        case 'play': {
            this.walk(2);
            const hop = Math.round(Math.abs(Math.sin(k * 0.5)) * 6);
            const ball = Math.round(Math.abs(Math.cos(k * 0.5)) * 10);
            return {frame: f.step, palette: PINK, dy: -hop, extras: [[BALL, this.front(-2, 6), GROUND + 20 - ball]]};
        }
        case 'adopt':
            return {frame: f.stand, palette: PINK, dy: -(k % 6 < 3 ? 3 : 0), extras: [[HEART, PIG_X + 4, GROUND - (k >> 2)], [HEART, PIG_X + 20, GROUND + 2 - (k >> 2)]]};
        case 'notice':
            // A message arrived: it pricks up and hops twice. Asleep, it only stirs.
            if (mood === 'sleeping' || mood === 'night') {
                return {frame: f.lie, palette: PINK, dy: k % 4 < 2 ? -1 : 0, extras: [[BANG, PIG_X + 22, GROUND + 4]]};
            }
            if (!ROAMING.has(mood)) {
                return null;
            }
            return {frame: f.look[k % f.look.length], palette: PINK, dy: k < 8 && k % 4 < 2 ? -3 : 0, extras: [[BANG, PIG_X + 22, GROUND - 2]]};
        default:
            return null;
        }
    }

    // pose decides the frame, palette, vertical offset and overlays of this tick.
    pose(mood) {
        const f = this.frames;
        const t = this.tick;
        if (this.reaction && t >= this.reaction.until) {
            this.reaction = null;
        }
        if (mood === 'dead') {
            return {frame: f.dead, palette: PALETTES.ghost, dy: 0, extras: [[HALO, PIG_X + 12, GROUND + 9 + (t % 16 < 8 ? 0 : 1)]]};
        }
        const reacting = this.reaction && this.reactionPose(mood);
        if (reacting) {
            return reacting;
        }
        if (mood === 'sleeping' || mood === 'night') {
            const rise = (t >> 1) % 12;
            return {frame: f.lie, palette: PINK, dy: t % 16 < 8 ? 0 : 1, extras: [[ZZZ, PIG_X + 24 + (rise >> 2), GROUND + 8 - rise]]};
        }
        if (mood === 'sick') {
            return {frame: t % 24 < 20 ? f.stand : f.blink, palette: PALETTES.sick, dy: 0, dx: t % 4 < 2 ? 0 : 1, extras: [[BANG, PIG_X + 15, GROUND - 1]]};
        }
        if (mood === 'sad') {
            const drop = (t >> 1) % 8;
            return {frame: f.sniff, palette: PINK, dy: 0, extras: [[TEAR, this.front(-9, 2), GROUND + 16 + drop]]};
        }
        if (this.target !== null && this.track && !this.paused) {
            const sniffing = this.chase(mood);
            if (sniffing) {
                return sniffing;
            }
            return {frame: (t >> 1) % 2 ? f.step : f.stand, palette: PINK, dy: 0, extras: []};
        }

        const bubble = mood === 'hungry' ? APPLE : null;
        if (t >= this.activity.until) {
            this.pickActivity(mood);
        }
        switch (this.activity.name) {
        case 'run': {
            this.walk(3);
            return {frame: t % 2 ? f.step : f.stand, palette: PINK, dy: t % 4 < 2 ? -2 : 0, extras: [], bubble};
        }
        case 'hop':
            return {frame: f.stand, palette: PINK, dy: -[0, 3, 5, 3][t % 4], extras: [], bubble};
        case 'trip': {
            const d = this.track.towards(this.pos, this.activity.to);
            if (Math.abs(d) < 8) {
                this.activity.until = t;
            } else {
                this.dir = Math.sign(d);
                this.walk(mood === 'happy' ? 4 : (lively(mood) ? 3 : 1));
            }
            return {frame: t % 2 ? f.step : f.stand, palette: PINK, dy: mood === 'happy' && t % 4 < 2 ? -1 : 0, extras: [], bubble};
        }
        case 'walk': {
            // Tired or hungry, it drags its feet.
            if (lively(mood)) {
                this.walk(2);
            } else if (t % 2 === 0) {
                this.walk(1);
            }
            return {frame: (t >> 1) % 2 ? f.step : f.stand, palette: PINK, dy: 0, extras: [], bubble};
        }
        case 'look':
            return {frame: f.look[(this.activity.until - t) % f.look.length], palette: PINK, dy: 0, extras: [], bubble};
        default:
            return {frame: t % 30 === 0 ? f.blink : f.stand, palette: PINK, dy: 0, extras: [], bubble: bubble || (mood === 'tired' ? ZZZ : null)};
        }
    }

    draw(ctx, mood) {
        this.tick++;
        const p = this.pose(mood);
        ctx.clearRect(0, 0, VIEW, VIEW);
        const x0 = PIG_X + (p.dx || 0);
        const y0 = GROUND + p.dy;
        for (let y = 0; y < SIZE; y++) {
            for (let x = 0; x < SIZE; x++) {
                const v = p.frame[y * SIZE + x];
                if (v) {
                    ctx.fillStyle = p.palette[v];
                    ctx.fillRect(x0 + (this.dir > 0 ? x : SIZE - 1 - x), y0 + y, 1, 1);
                }
            }
        }
        for (const [o, x, y, cut = 0] of p.extras) {
            drawOverlay(ctx, o, Math.round(x), Math.round(y), cut);
        }
        if (p.bubble && this.tick % 32 < 24) {
            drawBubble(ctx, p.bubble, this.dir > 0 ? x0 + 22 : x0 - 1, GROUND - 6);
        }
    }
}

// drawOverlay paints an overlay, leaving out its first `cut` columns (bitten apple).
function drawOverlay(ctx, o, x, y, cut) {
    for (const px of o.pixels) {
        if (px.x >= cut) {
            ctx.fillStyle = px.color;
            ctx.fillRect(x + px.x, y + px.y, 1, 1);
        }
    }
}

// A speech bubble with an overlay inside, e.g. an apple when hungry.
function drawBubble(ctx, o, x, y) {
    const w = o.width + 4;
    const h = o.height + 4;
    ctx.fillStyle = '#222034';
    ctx.fillRect(x, y, w, h);
    ctx.fillRect(x + 2, y + h, 2, 2);
    ctx.fillStyle = '#ffffff';
    ctx.fillRect(x + 1, y + 1, w - 2, h - 2);
    ctx.fillRect(x + 2, y + h - 1, 1, 1);
    drawOverlay(ctx, o, x + 2, y + 2, 0);
}
