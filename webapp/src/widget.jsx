import React from 'react';

import sheetUrl from '../../server/assets/piggy_look_anim-sheet.png';

import {fetchPet} from './api.js';
import {SIZE, buildFrames, gridFromRGBA} from './frames.js';
import Panel from './panel.jsx';
import Scene, {PIG_BOX, TICK_MS, VIEW} from './scene.js';
import {getPet, onReaction, setPet, subscribe} from './store.js';
import Track from './track.js';

const {useCallback, useEffect, useRef, useState, useSyncExternalStore} = React;

const SCALE = 2;
const BOX = VIEW * SCALE;
// The piggy goes to sniff the mouse when it rests this close to an edge (CSS
// pixels) for a moment, and loses interest if it stays put for long.
const MOUSE_REACH = 250;
const MOUSE_SETTLE_MS = 1000;
const MOUSE_BORED_MS = 45000;

function loadFrames() {
    return new Promise((resolve, reject) => {
        const img = new Image();
        img.onload = () => {
            const canvas = document.createElement('canvas');
            canvas.width = img.width;
            canvas.height = img.height;
            const ctx = canvas.getContext('2d');
            ctx.drawImage(img, 0, 0);
            const {data} = ctx.getImageData(0, 0, img.width, img.height);
            const sheet = [];
            for (let row = 0; row < img.height / SIZE; row++) {
                for (let col = 0; col < img.width / SIZE; col++) {
                    sheet.push(gridFromRGBA(data, img.width, col, row));
                }
            }
            resolve(buildFrames(sheet));
        };
        img.onerror = reject;
        img.src = sheetUrl;
    });
}

const newTrack = () => new Track(window.innerWidth, window.innerHeight, BOX);

// mouseTarget is the edge position to walk to, or null if the mouse is not resting near an edge.
function mouseTarget(mouse, track, now) {
    const still = now - mouse.at;
    if (!mouse.at || still < MOUSE_SETTLE_MS || still > MOUSE_BORED_MS) {
        return null;
    }
    const near = track.nearest(mouse.x, mouse.y);
    return near.distance <= MOUSE_REACH ? near.s : null;
}

export default function Widget({t, locale}) {
    const pet = useSyncExternalStore(subscribe, getPet);
    const boxRef = useRef(null);
    const canvasRef = useRef(null);
    const sceneRef = useRef(null);
    const moodRef = useRef('ok');
    const [anchor, setAnchor] = useState(null);
    const [error, setError] = useState(false);

    moodRef.current = pet?.mood ?? 'ok';

    useEffect(() => {
        fetchPet().then(setPet, () => setError(true));
    }, []);

    useEffect(() => {
        let timer;
        let cancelled = false;
        let track = newTrack();
        const mouse = {x: 0, y: 0, at: 0};
        const onMouse = (e) => Object.assign(mouse, {x: e.clientX, y: e.clientY, at: Date.now()});
        const onLeave = (e) => {
            if (!e.relatedTarget) {
                mouse.at = 0;
            }
        };
        const onResize = () => {
            track = newTrack();
            if (sceneRef.current) {
                sceneRef.current.track = track;
            }
        };
        window.addEventListener('mousemove', onMouse, {passive: true});
        document.addEventListener('mouseout', onLeave);
        window.addEventListener('resize', onResize);

        loadFrames().then((frames) => {
            if (cancelled) {
                return;
            }
            const scene = new Scene(frames, SCALE);
            scene.track = track;
            sceneRef.current = scene;
            const ctx = canvasRef.current.getContext('2d');
            timer = setInterval(() => {
                scene.target = mouseTarget(mouse, track, Date.now());
                scene.draw(ctx, moodRef.current);
                const {left, top, rotate} = track.at(scene.pos);
                boxRef.current.style.transform = `translate(${left}px, ${top}px) rotate(${rotate}deg)`;
            }, TICK_MS);
        }, () => setError(true));
        const offReaction = onReaction((action) => sceneRef.current?.react(action));
        return () => {
            cancelled = true;
            clearInterval(timer);
            offReaction();
            window.removeEventListener('mousemove', onMouse);
            document.removeEventListener('mouseout', onLeave);
            window.removeEventListener('resize', onResize);
        };
    }, []);

    // The piggy stands still while its panel is open, so the panel stays next to it.
    const toggle = () => {
        const scene = sceneRef.current;
        if (!scene) {
            return;
        }
        if (anchor) {
            scene.paused = false;
            setAnchor(null);
            return;
        }
        scene.paused = true;
        setAnchor({...scene.track.at(scene.pos), size: BOX});
    };
    const close = useCallback(() => {
        if (sceneRef.current) {
            sceneRef.current.paused = false;
        }
        setAnchor(null);
    }, []);

    const hit = PIG_BOX;
    return (
        <>
            <div
                ref={boxRef}
                className='piggy-widget'
                style={{width: BOX, height: BOX}}
            >
                <canvas
                    ref={canvasRef}
                    className='piggy-canvas'
                    width={VIEW}
                    height={VIEW}
                    style={{width: BOX, height: BOX}}
                />
                <div
                    className='piggy-hit'
                    role='button'
                    tabIndex={0}
                    style={{left: hit.x * SCALE, top: hit.y * SCALE, width: hit.width * SCALE, height: hit.height * SCALE}}
                    aria-label={pet?.name}
                    title={pet ? `${pet.name} · ${t(`mood_${pet.mood}`)}` : ''}
                    onClick={toggle}
                    onKeyDown={(e) => (e.key === 'Enter' || e.key === ' ') && toggle()}
                />
            </div>
            {anchor && (
                <Panel
                    pet={pet}
                    error={error}
                    t={t}
                    locale={locale}
                    anchor={anchor}
                    onClose={close}
                />
            )}
        </>
    );
}
