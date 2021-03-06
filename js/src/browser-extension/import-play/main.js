import App from './app.svelte';
import Interact from "../../components/interact/interact_v2.svelte";
import Interact2 from "./routes/sr.svelte";
import Banner from "../../components/banner/banner.svelte";


new Banner({
	target: document.querySelector("#notification-center"),
});

// TODO if I copy the functionality, I can control overtime.
// Currently it breaks
new Interact({
	target: document.querySelector("#play-screen"),
});

new Interact2({
	target: document.querySelector("#play-screen-2"),
});

new App({
	target: document.querySelector("#list-info"),
});
