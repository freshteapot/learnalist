import { writable } from 'svelte/store';
import { getConfiguration, saveConfiguration, clearConfiguration, KeyUserAuthentication } from '../configuration.js';

function loginHelperSingleton() {
    const defaultRedirectURL = '/';
    let obj = {
        redirectURL: defaultRedirectURL,
        loggedIn: (() => {
            const auth = getConfiguration(KeyUserAuthentication);
            return auth ? true : false;
        })()
    }

    const { subscribe, set, update } = writable(obj);

    return {
        subscribe,

        login: ((session) => {
            saveConfiguration(KeyUserAuthentication, session.token);
            update(n => {
                n.loggedIn = true;
                return n;
            });
        }),

        logout: () => {
            clearConfiguration();
            update(n => {
                n.loggedIn = false;
                return n;
            });
        },

        redirectURLAfterLogin: (redirectURL) => {
            if (isStringEmpty(redirectURL)) {
                redirectURL = defaultRedirectURL;
            }

            update(n => {
                n.redirectURL = redirectURL;
                return n;
            });
        }
    };
}

const loginHelper = loginHelperSingleton();

export {
    loginHelper
}
