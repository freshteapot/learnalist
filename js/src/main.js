import { KeySettingsInstallDefaults, get as cacheGet, clear as clearCache } from './cache.js';
import Banner from './components/banner/banner.svelte';
import LoginHeader from './components/login_header.svelte';
import UserLogin from './components/user_login.svelte';

const installed = cacheGet(KeySettingsInstallDefaults, null)
if (installed === null) {
    clearCache();
}

// TODO setup
customElements.define('login-header', LoginHeader);
customElements.define('user-login', UserLogin);
customElements.define('notification-center', Banner);

