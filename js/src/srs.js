
import SpacedRepetition from "./components/spaced-repetition/sample.svelte";

// Actual app to handle the interactions
let app;
const el = document.querySelector("#play")
if (el) {
    app = new SpacedRepetition({
        target: el,
    });
}

export default app;
