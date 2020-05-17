
import Menu from './components/interact/menu.wc.svelte';
import Interact from "./components/interact/interact.svelte";

// Webcomponent
customElements.define('v1-menu', Menu);

// Actual app to handle the interactions
let app;
const el = document.querySelector("#play-screen")
if (el) {
    app = new Interact({
        target: el,
    });
}

export default app;
