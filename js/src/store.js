import Cookies from './js.cookie.js'
import { writable } from 'svelte/store';
import { notifications } from './notifications.js';

const ID_LOGGED_IN_KEY = 'x-authentication-bearer';

const count = writable(0);

// Link any component to be able to notify the banner component
const notify = (level, message) => {
    notifications.add(level, message);
}

const loggedIn = () => {
    console.log("am I logged in");
    console.log("Should I check local storage");
    let item = Cookies.get(ID_LOGGED_IN_KEY);
    if (!item) {
        return false;
    }
    return true;
}

const logout = (redirect) => {
    console.log("I want to be logged out.")
    Cookies.remove(ID_LOGGED_IN_KEY);
    localStorage.clear();
    if (redirect === "#") {
        return;
    }
    if (!redirect) {
        redirect = '/welcome.html';
    }
    window.location = redirect;
}

const login = (redirect) => {
    if (!redirect) {
        redirect = '/welcome.html';
    }
    window.location = redirect;
}

export {
    count,
    loggedIn,
    logout,
    login,
    notify,
    notifications,
}
