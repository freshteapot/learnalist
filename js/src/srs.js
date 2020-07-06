
import SpacedRepetition from "./components/interact/spaced_repetition/view.svelte";

// Actual app to handle the interactions
let app;
const el = document.querySelector("#play")
if (el) {
    app = new SpacedRepetition({
        target: el,
    });
}

export default app;
