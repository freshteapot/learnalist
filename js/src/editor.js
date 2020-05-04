import App from './editor/App.svelte';
import cache from './editor/lib/cache.js';

if (!localStorage.hasOwnProperty(cache.keys['settings.install.defaults'])) {
    cache.clear();
    window.location = location.origin + location.pathname;
}

// Specific for the chrome extension
if (window.location.protocol === 'chrome-extension:') {
    let last = cache.get(cache.keys['last.screen']);
    console.log(last);
    if (last) {
        if (last !== window.location.hash) {
            history.replaceState(undefined, undefined, last);
            window.dispatchEvent(new Event('hashchange'));
        }
    }
}

if (window.location.protocol !== 'chrome-extension:') {
    let last = cache.get(cache.keys['last.screen']);
    if (last) {
        if (last !== window.location.hash) {
            history.replaceState(undefined, undefined, (location.origin + location.pathname + last));
            window.dispatchEvent(new Event('hashchange'));
        }
    }
}


var app = new App({
    target: document.querySelector("#list-info")
});

export default app;
