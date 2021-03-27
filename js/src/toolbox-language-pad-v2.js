
import Experience from "./toolbox/language-pad/v2.svelte";

// Actual app to handle the interactions
let app;
const el = document.querySelector("#main-panel")
if (el) {
    app = new Experience({
        target: el,
    });
}

export default app;
