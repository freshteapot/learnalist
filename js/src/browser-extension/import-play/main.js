import App from './App.svelte';
import Interact from "../../components/interact/interact_v2.svelte";

new Interact({
	target: document.querySelector("#play-screen"),
});

new App({
	target: document.querySelector("#list-info"),
});


