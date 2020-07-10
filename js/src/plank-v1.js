
import HumblePlank from "./plank/index.svelte";

// Actual app to handle the interactions
let app;
const el = document.querySelector("#list-info")
if (el) {
    app = new HumblePlank({
        target: el,
    });
}

export default app;
