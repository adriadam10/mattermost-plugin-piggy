// Where the piggy's panel goes: next to the piggy, on the inner side of the
// window edge it is walking on, and always inside the window.

// Gap between the panel and the piggy or the window border, in CSS pixels.
const GAP = 8;

const clampTo = (v, lo, hi) => Math.max(lo, Math.min(hi, v));

// placement puts the panel next to the piggy, on the inner side of the edge it is on.
export function placement(anchor, width, height, viewW, viewH) {
    const {left, top, size, edge} = anchor;
    let x = left + (size - width) / 2;
    let y = top + (size - height) / 2;
    switch (edge) {
    case 'bottom':
        y = top - height - GAP;
        break;
    case 'top':
        y = top + size + GAP;
        break;
    case 'right':
        x = left - width - GAP;
        break;
    default:
        x = left + size + GAP;
    }
    return {
        left: clampTo(x, GAP, viewW - width - GAP),
        top: clampTo(y, GAP, viewH - height - GAP),
    };
}
