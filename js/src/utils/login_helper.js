import { writable } from 'svelte/store';
import { save as saveCache, get as cacheGet } from './storage.js';
import { KeyUserAuthentication } from '../cache.js';

function loginHelperSingleton() {
    const defaultRedirectURL = '/';
    let obj = {
        redirectURL: defaultRedirectURL,
        loggedIn: (() => {
            const auth = cacheGet(KeyUserAuthentication);
            return auth ? true : false;
        })()
    }

    const { subscribe, set, update } = writable(obj);

    return {
        subscribe,

        login: ((session) => {
            saveCache(KeyUserAuthentication, session.token);
            update(n => {
                n.loggedIn = true;
                return n;
            });
        }),

        logout: () => {
            cache.clear()
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
