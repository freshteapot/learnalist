import { KeySettingsInstallDefaults, getConfiguration, clearConfiguration } from './configuration.js';
import Banner from './components/banner/banner.svelte';
import LoginHeader from './components/login_header.svelte';
import UserLogin from './components/user_login.svelte';
import HumblePlank from "./plank/index.svelte";
import { loggedIn, notify } from "./shared.js";


// The crudest attempt to see if we have setup the site with configuration in localstorage
const installed = getConfiguration(KeySettingsInstallDefaults, null)
if (installed === null) {
    clearConfiguration();
}

function connect(id, element) {
    const el = document.querySelector(id);
    if (el) {
        new element({
            target: el,
        });
    }
}

connect("#notification-center", Banner)
connect("#user-login", UserLogin)
connect("#login-header", LoginHeader)



let app;
const querystring = window.location.search;
const searchParams = new URLSearchParams(querystring);


if (loggedIn()) {
    const el = document.querySelector("#main-panel")
    if (el) {
        app = new HumblePlank({
            target: el,
        });
    }
} else {
    if (!searchParams.has("redirect")) {
        window.location.href = window.location.href + "?redirect=plank2.html"
    }

    shared.notify("info", "Please login")
    connect("#plank-login", UserLogin)
}
export default app;
