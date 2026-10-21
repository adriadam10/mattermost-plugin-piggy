import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {test} from 'node:test';
import {inflateSync} from 'node:zlib';

import {BODY, DARK, SIZE, buildFrames, findEye, gridFromRGBA, overlay} from './frames.js';

// Minimal decoder for the sheet (8-bit RGBA, non interlaced), so the test reads the real file.
function decodePNG(buf) {
    const width = buf.readUInt32BE(16);
    const height = buf.readUInt32BE(20);
    const chunks = [];
    for (let i = 8; i < buf.length;) {
        const len = buf.readUInt32BE(i);
        if (buf.toString('ascii', i + 4, i + 8) === 'IDAT') {
            chunks.push(buf.subarray(i + 8, i + 8 + len));
        }
        i += len + 12;
    }
    const raw = inflateSync(Buffer.concat(chunks));
    const stride = width * 4;
    const out = new Uint8Array(height * stride);
    for (let y = 0; y < height; y++) {
        const filter = raw[y * (stride + 1)];
        for (let x = 0; x < stride; x++) {
            const cur = raw[y * (stride + 1) + 1 + x];
            const a = x >= 4 ? out[y * stride + x - 4] : 0;
            const b = y > 0 ? out[(y - 1) * stride + x] : 0;
            const c = x >= 4 && y > 0 ? out[(y - 1) * stride + x - 4] : 0;
            const p = a + b - c;
            const paeth = Math.abs(p - a) <= Math.abs(p - b) && Math.abs(p - a) <= Math.abs(p - c) ? a : (Math.abs(p - b) <= Math.abs(p - c) ? b : c);
            out[y * stride + x] = (cur + [0, a, b, (a + b) >> 1, paeth][filter]) & 0xff;
        }
    }
    return {width, data: out};
}

const {width, data} = decodePNG(readFileSync(new URL('../../server/assets/piggy_look_anim-sheet.png', import.meta.url)));
const sheet = [];
for (let row = 0; row < 3; row++) {
    for (let col = 0; col < 3; col++) {
        sheet.push(gridFromRGBA(data, width, col, row));
    }
}
const frames = buildFrames(sheet);
const count = (g, v) => g.filter((p) => p === v).length;

test('reads the three sheet colours', () => {
    assert.ok(count(frames.stand, BODY) > 300);
    assert.deepEqual(findEye(frames.stand), {x: 23, y: 13});
});

test('blinking and lying close the eye', () => {
    assert.equal(findEye(frames.blink), null);
    assert.equal(findEye(frames.lie), null);
});

test('lying keeps the body and drops it onto the ground', () => {
    const top = (g) => g.findIndex((p) => p !== 0) / SIZE | 0;
    assert.equal(top(frames.lie), top(frames.stand) + 4);
});

test('walking moves the legs, not the body', () => {
    const changed = frames.stand.reduce((n, p, i) => n + (p !== frames.step[i]), 0);
    assert.ok(changed > 0 && changed < 40, `${changed} pixels changed`);
});

test('dead is upside down with X eyes', () => {
    assert.equal(count(frames.dead, DARK), count(frames.stand, DARK) + 4);
});

test('overlays parse to pixels', () => {
    const heart = overlay('heart');
    assert.equal(heart.width, 7);
    assert.ok(heart.pixels.every((p) => p.color === '#e43b44'));
});
