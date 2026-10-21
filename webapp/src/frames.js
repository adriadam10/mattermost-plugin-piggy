// Frames of the piggy as 32x32 grids of palette indices, derived from the CC0
// sprite sheet: it only has a "look around" animation facing right, so walking,
// sleeping and dying are small pixel edits of its first frame.

export const SIZE = 32;
export const EMPTY = 0;
export const BODY = 1;
export const SHADE = 2;
export const DARK = 3;

// The sheet's three colours, in palette order.
const SHEET_COLORS = ['238,195,154', '217,160,102', '143,86,59'];

export const PALETTES = {
    pink: [null, '#ffb7c5', '#e8879f', '#8a3b52'],
    sick: [null, '#c9d9a0', '#9fb575', '#5d6b3a'],
    ghost: [null, '#d9d9d9', '#a8a8a8', '#555555'],
};

// gridFromRGBA reads frame (col, row) of an RGBA pixel buffer `width` pixels wide.
export function gridFromRGBA(data, width, col, row) {
    const grid = new Uint8Array(SIZE * SIZE);
    for (let y = 0; y < SIZE; y++) {
        for (let x = 0; x < SIZE; x++) {
            const i = (((row * SIZE) + y) * width + (col * SIZE) + x) * 4;
            if (data[i + 3] === 0) {
                continue;
            }
            grid[y * SIZE + x] = SHEET_COLORS.indexOf(`${data[i]},${data[i + 1]},${data[i + 2]}`) + 1;
        }
    }
    return grid;
}

const get = (g, x, y) => (x < 0 || y < 0 || x >= SIZE || y >= SIZE ? EMPTY : g[y * SIZE + x]);
const set = (g, x, y, v) => {
    if (x >= 0 && y >= 0 && x < SIZE && y < SIZE) {
        g[y * SIZE + x] = v;
    }
};

// The eye is the only dark pixel above the hooves.
export function findEye(g) {
    for (let y = 0; y < 24; y++) {
        for (let x = 0; x < SIZE; x++) {
            if (get(g, x, y) === DARK) {
                return {x, y};
            }
        }
    }
    return null;
}

function blink(g) {
    const out = g.slice();
    const eye = findEye(g);
    if (eye) {
        set(out, eye.x, eye.y, SHADE);
    }
    return out;
}

// Moves the pixels of a rectangle dx columns sideways.
function shift(g, x0, x1, y0, y1, dx) {
    const out = g.slice();
    for (let y = y0; y <= y1; y++) {
        for (let x = x0; x <= x1; x++) {
            set(out, x, y, EMPTY);
        }
    }
    for (let y = y0; y <= y1; y++) {
        for (let x = x0; x <= x1; x++) {
            const v = get(g, x, y);
            if (v !== EMPTY) {
                set(out, x + dx, y, v);
            }
        }
    }
    return out;
}

// Second walking frame: legs (rows 23-27) swung, the back one forward and the front one back.
function step(g) {
    return shift(shift(g, 0, 5, 23, 27, 1), 17, 23, 23, 27, -1);
}

// Lying down: the body drops onto the folded legs and the eye closes into a line.
function lie(g) {
    const out = new Uint8Array(SIZE * SIZE);
    for (let y = 0; y < 24; y++) {
        for (let x = 0; x < SIZE; x++) {
            set(out, x, y + 4, get(g, x, y));
        }
    }
    const eye = findEye(out);
    if (eye) {
        for (const dx of [-1, 0, 1]) {
            set(out, eye.x + dx, eye.y, SHADE);
        }
    }
    return out;
}

// Dead: upside down with the legs up and the eye as an X.
function dead(g) {
    const out = new Uint8Array(SIZE * SIZE);
    for (let y = 0; y < SIZE; y++) {
        for (let x = 0; x < SIZE; x++) {
            set(out, x, 33 - y, get(g, x, y));
        }
    }
    // Looked up before flipping: upside down, the hooves come first.
    const eye = findEye(g);
    if (eye) {
        for (const [dx, dy] of [[0, 0], [-1, -1], [1, -1], [-1, 1], [1, 1]]) {
            set(out, eye.x + dx, 33 - eye.y + dy, DARK);
        }
    }
    return out;
}

// frames builds every pose from the nine sheet frames, in reading order.
export function buildFrames(sheet) {
    const stand = sheet[0];
    return {
        stand,
        blink: blink(stand),
        step: step(stand),
        lie: lie(stand),
        dead: dead(stand),
        // Head turned towards the viewer, snout down: eating and moping.
        sniff: sheet[6],
        look: sheet,
    };
}

// Small overlays as text, one character per pixel and a colour per character.
const OVERLAYS = {
    heart: {
        colors: {X: '#e43b44'},
        rows: ['.XX.XX.', 'XXXXXXX', 'XXXXXXX', '.XXXXX.', '..XXX..', '...X...'],
    },
    zzz: {
        colors: {X: '#3978a8'},
        rows: ['XXXXX', '...X.', '..X..', '.X...', 'XXXXX'],
    },
    apple: {
        colors: {R: '#e43b44', H: '#ffaaaa', G: '#3e8948'},
        rows: ['...G...', '..G....', '.RRRRR.', 'RRHRRRR', 'RHRRRRR', 'RRRRRRR', '.RRRRR.'],
    },
    ball: {
        colors: {W: '#ffffff', B: '#3978a8'},
        rows: ['.BBBB.', 'BWWBBB', 'BWBBBB', 'BBBBWB', 'BBBWWB', '.BBBB.'],
    },
    tear: {
        colors: {T: '#63c6ff'},
        rows: ['.T', 'TT', 'TT'],
    },
    halo: {
        colors: {Y: '#feae34'},
        rows: ['.YYYYYYY.', 'Y.......Y', '.YYYYYYY.'],
    },
    bang: {
        colors: {X: '#e43b44'},
        rows: ['XX', 'XX', 'XX', 'XX', '..', 'XX'],
    },
};

export function overlay(name) {
    const {colors, rows} = OVERLAYS[name];
    const pixels = [];
    rows.forEach((row, y) => [...row].forEach((ch, x) => {
        if (colors[ch]) {
            pixels.push({x, y, color: colors[ch]});
        }
    }));
    return {width: rows[0].length, height: rows.length, pixels};
}
