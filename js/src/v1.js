import LoginHeader from './components/login_header.svelte';
import Banner from './components/banner/banner.svelte';

import Slideshow from './components/v1/Slideshow.svelte';
import Menu from './components/v1/Menu.svelte';

customElements.define('login-header', LoginHeader);
customElements.define('notification-center', Banner);
customElements.define('v1-menu', Menu);
customElements.define('v1-slideshow', Slideshow);
