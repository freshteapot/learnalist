import { KeySettingsInstallDefaults, get as cacheGet, clear as clearCache } from './cache.js';
import Banner from './components/banner/banner.svelte';
import LoginHeader from './components/login_header.svelte';
import UserLogin from './components/user_login.svelte';

const installed = cacheGet(KeySettingsInstallDefaults, null)
if (installed === null) {
    clearCache();
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
