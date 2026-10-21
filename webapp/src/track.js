// The piggy walks round the edges of the window, feet against the edge: along
// the bottom to the right, up the right side, along the top to the left and down
// the left side. A position is a distance along that loop in CSS pixels, and the
// piggy's box is a square of `size` pixels.

// Rotation of the box on each edge, so that "forward" is the way round the loop
// and its feet touch the edge.
const EDGES = ['bottom', 'right', 'top', 'left'];
const ROTATION = {bottom: 0, right: -90, top: 180, left: 90};

export default class Track {
    constructor(width, height, size) {
        this.size = size;
        this.lengths = [width - size, height - size, width - size, height - size].map((l) => Math.max(0, l));
        this.width = width;
        this.height = height;
        this.length = this.lengths.reduce((a, b) => a + b, 0);
    }

    wrap(s) {
        return this.length ? ((s % this.length) + this.length) % this.length : 0;
    }

    // at is where the box goes for position s: its top left corner and rotation.
    at(s) {
        if (!this.length) {
            return {edge: 'bottom', left: 0, top: Math.max(0, this.height - this.size), rotate: 0};
        }
        let rest = this.wrap(s);
        let i = 0;
        while (i < 3 && rest >= this.lengths[i]) {
            rest -= this.lengths[i];
            i++;
        }
        const {width: w, height: h, size} = this;
        const edge = EDGES[i];
        switch (edge) {
        case 'bottom':
            return {edge, left: rest, top: h - size, rotate: ROTATION[edge]};
        case 'right':
            return {edge, left: w - size, top: h - size - rest, rotate: ROTATION[edge]};
        case 'top':
            return {edge, left: w - size - rest, top: 0, rotate: ROTATION[edge]};
        default:
            return {edge, left: 0, top: rest, rotate: ROTATION[edge]};
        }
    }

    // nearest is the position of the edge point closest to (x, y), and how far it is.
    nearest(x, y) {
        const {width: w, height: h, size} = this;
        const half = size / 2;
        const clampX = Math.max(half, Math.min(w - half, x));
        const clampY = Math.max(half, Math.min(h - half, y));
        const [lb, lr, lt] = this.lengths;
        const candidates = [
            {s: clampX - half, distance: h - y},
            {s: lb + (h - half - clampY), distance: w - x},
            {s: lb + lr + (w - half - clampX), distance: y},
            {s: lb + lr + lt + (clampY - half), distance: x},
        ];
        return candidates.reduce((best, c) => (c.distance < best.distance ? c : best));
    }

    // towards is the signed shortest way round the loop from a to b.
    towards(a, b) {
        const d = this.wrap(b - a);
        return d > this.length / 2 ? d - this.length : d;
    }
}
