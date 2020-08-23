
import SpacedRepetition from "./spaced_repetition/app.svelte";

// Actual app to handle the interactions
let app;
const el = document.querySelector("#main-panel")
if (el) {
    app = new SpacedRepetition({
        target: el,
    });
}

export default app;
