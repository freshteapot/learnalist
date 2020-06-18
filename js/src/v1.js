
import Menu from './components/interact/menu.wc.svelte';
import Interact from "./components/interact/interact_v1.svelte";

// Webcomponent
customElements.define('interact-menu', Menu);

// Actual app to handle the interactions
let app;
const el = document.querySelector("#play-screen")
if (el) {
    app = new Interact({
        target: el,
    });
}

export default app;
