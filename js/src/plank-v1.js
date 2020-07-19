import HumblePlank from "./plank/index.svelte";
let app;
const el = document.querySelector("#main-panel")
if (el) {
    app = new HumblePlank({
        target: el,
    });
}

export default app;
