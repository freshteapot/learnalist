import { notifications } from './notifications.js';
import { KeyUserAuthentication } from './configuration.js';
import * as api from './api.js';

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
    api
}
