import { KeySettingsInstallDefaults, getConfiguration, clearConfiguration } from './configuration.js';
import Banner from './components/banner/banner.svelte';
import LoginHeader from './components/login_header.svelte';
import UserLogin from './components/user_login.svelte';

//// The crudest attempt to see if we have setup the site with configuration in localstorage
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
