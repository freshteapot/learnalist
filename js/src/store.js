import { notifications } from './notifications.js';
import { KeyUserAuthentication } from './cache.js';

// Link any component to be able to notify the banner component
const notify = (level, message) => {
    notifications.add(level, message);
}

const clearNotification = () => {
    notifications.clear();
}

const loggedIn = () => {
    return localStorage.hasOwnProperty(KeyUserAuthentication);
}

/*
const logout = (redirect) => {
    console.log("I want to be logged out.")
    // TODO how to make this work when I dont know the domain.
    const apiServer = document.querySelector('meta[name="api.server"]');
    // TODO this will need to know if its secure.
    Cookies.remove(ID_LOGGED_IN_KEY, { path: '/', domain: `.${apiServer}` });
    localStorage.clear();
    if (redirect === "#") {
        return;
    }
    if (!redirect) {
        redirect = '/welcome.html';
    }
    window.location = redirect;
}
*/

const login = (redirect) => {
    if (!redirect) {
        redirect = '/welcome.html';
    }
    window.location = redirect;
}

export {
    login,
    loggedIn,
    notify,
    notifications,
    clearNotification,
}
