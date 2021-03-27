
import Experience from "undefined.svelte";

// Actual app to handle the interactions
let app;
const el = document.querySelector("#main-panel")
if (el) {
    app = new Experience({
        target: el,
    });
}

export default app;
