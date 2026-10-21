import assert from 'node:assert/strict';
import {test} from 'node:test';

import {placement} from './placement.js';

const W = 1200;
const H = 800;
const size = 112;

test('opens towards the inside of each edge', () => {
    assert.deepEqual(placement({edge: 'bottom', left: 500, top: H - size, size}, 280, 300, W, H), {left: 416, top: H - size - 300 - 8});
    assert.equal(placement({edge: 'top', left: 500, top: 0, size}, 280, 300, W, H).top, size + 8);
    assert.equal(placement({edge: 'right', left: W - size, top: 300, size}, 280, 300, W, H).left, W - size - 280 - 8);
    assert.equal(placement({edge: 'left', left: 0, top: 300, size}, 280, 300, W, H).left, size + 8);
});

test('stays inside the window near the corners', () => {
    assert.deepEqual(placement({edge: 'bottom', left: 0, top: H - size, size}, 280, 300, W, H).left, 8);
    assert.deepEqual(placement({edge: 'right', left: W - size, top: 0, size}, 280, 300, W, H).top, 8);
});
