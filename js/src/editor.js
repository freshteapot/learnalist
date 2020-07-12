import App from './editor/App.svelte';
import { getConfiguration, KeyLastScreen } from './configuration.js';

let last = getConfiguration(KeyLastScreen, undefined);
if (last) {
    if (last !== window.location.hash) {
        history.replaceState(undefined, undefined, (location.origin + location.pathname + last));
        window.dispatchEvent(new Event('hashchange'));
    }
}



var app = new App({
    target: document.querySelector("#list-info")
});

export default app;
