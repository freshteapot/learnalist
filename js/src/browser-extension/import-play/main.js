import App from './app.svelte';
import Interact from "../../components/interact/interact_v2.svelte";
import InteractSR from "./routes/sr.svelte";
import Banner from "../../components/banner/banner.svelte";


new Banner({
	target: document.querySelector("#notification-center"),
});

new Interact({
	target: document.querySelector("#play-screen"),
});

new InteractSR({
	target: document.querySelector("#play-screen-sr"),
});

new App({
	target: document.querySelector("#list-info"),
});
