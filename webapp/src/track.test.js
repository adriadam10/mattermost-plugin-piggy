import assert from 'node:assert/strict';
import {test} from 'node:test';

import Track from './track.js';

const track = new Track(1000, 600, 100);

test('goes round all four edges', () => {
    assert.equal(track.length, 900 + 500 + 900 + 500);
    assert.deepEqual(track.at(0), {edge: 'bottom', left: 0, top: 500, rotate: 0});
    assert.deepEqual(track.at(900), {edge: 'right', left: 900, top: 500, rotate: -90});
    assert.deepEqual(track.at(1400), {edge: 'top', left: 900, top: 0, rotate: 180});
    assert.deepEqual(track.at(2300), {edge: 'left', left: 0, top: 0, rotate: 90});
    assert.deepEqual(track.at(2800), track.at(0));
    assert.deepEqual(track.at(-100), track.at(2700));
});

test('nearest edge point', () => {
    assert.deepEqual(track.nearest(300, 590), {s: 250, distance: 10});
    assert.equal(track.at(track.nearest(990, 300).s).edge, 'right');
    assert.equal(track.at(track.nearest(500, 5).s).edge, 'top');
    assert.equal(track.at(track.nearest(3, 200).s).edge, 'left');
});

test('shortest way round', () => {
    assert.equal(track.towards(100, 300), 200);
    assert.equal(track.towards(100, 2700), -200);
});

test('tiny windows do not break', () => {
    const tiny = new Track(50, 50, 100);
    assert.equal(tiny.length, 0);
    assert.deepEqual(tiny.at(10).edge, 'bottom');
});
