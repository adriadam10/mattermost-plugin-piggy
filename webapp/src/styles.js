// 8-bit look: square corners, hard offset shadows, segmented bars. Colours come
// from the Mattermost theme variables, so it follows light and dark themes.
export const STYLES = `
.piggy-widget {
    position: fixed;
    left: 0;
    top: 0;
    z-index: 100;
    pointer-events: none;
}
.piggy-canvas {
    image-rendering: pixelated;
    display: block;
}
.piggy-hit {
    position: absolute;
    pointer-events: auto;
    cursor: pointer;
    touch-action: none;
}
.piggy-hit:focus-visible {
    outline: 2px dashed var(--button-bg);
}
.piggy-panel {
    position: fixed;
    z-index: 101;
    width: 260px;
    max-height: calc(100vh - 16px);
    overflow-y: auto;
    font-family: ui-monospace, "Cascadia Mono", Menlo, Consolas, monospace;
    padding: 10px;
    pointer-events: auto;
    background: var(--center-channel-bg);
    color: var(--center-channel-color);
    border: 3px solid var(--center-channel-color);
    box-shadow: 5px 5px 0 rgba(var(--center-channel-color-rgb), 0.3);
    font-size: 12px;
    line-height: 1.4;
}
.piggy-head {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 8px;
    text-transform: uppercase;
}
.piggy-head strong {
    font-size: 13px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
.piggy-mood {
    padding: 0 4px;
    border: 2px solid currentColor;
    font-size: 10px;
    white-space: nowrap;
}
.piggy-mood-happy, .piggy-mood-ok { color: var(--online-indicator); }
.piggy-mood-hungry, .piggy-mood-sad, .piggy-mood-tired { color: var(--away-indicator); }
.piggy-mood-sick, .piggy-mood-dead { color: var(--dnd-indicator); }
.piggy-mood-sleeping, .piggy-mood-night { color: rgba(var(--center-channel-color-rgb), 0.64); }
.piggy-close {
    margin-left: auto;
    border: 0;
    background: none;
    color: inherit;
    font: inherit;
    font-size: 18px;
    line-height: 1;
    cursor: pointer;
}
.piggy-stat {
    display: flex;
    align-items: center;
    gap: 8px;
    margin: 4px 0;
}
.piggy-stat-label {
    width: 64px;
    flex: none;
    text-transform: uppercase;
    font-size: 10px;
}
.piggy-bar {
    display: flex;
    gap: 2px;
    padding: 2px;
    border: 2px solid var(--center-channel-color);
    flex: 1;
}
.piggy-bar span {
    flex: 1;
    height: 8px;
    background: rgba(var(--center-channel-color-rgb), 0.1);
}
.piggy-bar-good span.on { background: var(--online-indicator); }
.piggy-bar-low span.on { background: var(--away-indicator); }
.piggy-bar-bad span.on { background: var(--dnd-indicator); }
.piggy-actions {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 6px;
    margin-top: 10px;
}
.piggy-btn {
    padding: 5px 4px;
    border: 2px solid var(--center-channel-color);
    background: var(--button-bg);
    color: var(--button-color);
    box-shadow: 3px 3px 0 var(--center-channel-color);
    font: inherit;
    font-weight: bold;
    text-transform: uppercase;
    cursor: pointer;
}
.piggy-btn:active:not(:disabled) {
    transform: translate(3px, 3px);
    box-shadow: none;
}
.piggy-btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
}
.piggy-btn-wide {
    grid-column: 1 / -1;
}
.piggy-notice {
    margin: 8px 0 0;
    font-weight: bold;
}
.piggy-rip {
    padding: 8px;
    text-align: center;
    border: 2px dashed rgba(var(--center-channel-color-rgb), 0.5);
}
.piggy-rip-title {
    font-size: 18px;
    font-weight: bold;
    letter-spacing: 2px;
}
.piggy-log {
    margin: 10px 0 0;
    padding: 6px 0 0;
    list-style: none;
    border-top: 2px dotted rgba(var(--center-channel-color-rgb), 0.3);
    font-size: 11px;
}
.piggy-log li {
    display: flex;
    justify-content: space-between;
    gap: 8px;
}
.piggy-log li span {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
.piggy-log time {
    flex: none;
    color: rgba(var(--center-channel-color-rgb), 0.64);
}
.piggy-foot {
    display: flex;
    flex-direction: column;
    margin-top: 8px;
    font-size: 10px;
    color: rgba(var(--center-channel-color-rgb), 0.64);
}
/* The phone layout has no room to spare; the slash command still works. */
@media (max-width: 768px) {
    .piggy-widget { display: none; }
}
`;
