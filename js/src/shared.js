import { notifications } from './notifications.js';
import { KeyUserAuthentication } from './configuration.js';
import * as api from './api.js';

// Link any component to be able to notify the banner component
const notify = (level, message, sticky) => {
    notifications.add(level, message, sticky);
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
