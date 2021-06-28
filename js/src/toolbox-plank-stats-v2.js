// Auto generated from rollup.config.toolbox.js
import Experience from "./plank/stats/v2/v2.svelte";

// Actual app to handle the interactions
let app;
const el = document.querySelector("#main-panel")
if (el) {
    app = new Experience({
        target: el,
    });
}

export default app;