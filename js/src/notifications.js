import { writable } from 'svelte/store';
import { rm as rmCache, save as saveCache, get as cacheGet, KeyNotifications } from './cache.js';

const data = {
    level: "",
    message: "",
}

const emptyData = JSON.parse(JSON.stringify(data));
let liveData = JSON.parse(JSON.stringify(data));


const storedData = cacheGet(KeyNotifications, null);

if (storedData !== null) {
    liveData = storedData;
}

const { subscribe, update, set } = writable(liveData);

function wrapper() {
    return {
        subscribe,

        add: (level, message) => {
            update(notification => {
                notification.level = level;
                notification.message = message;
                saveCache(KeyNotifications, notification)
                return notification;
            });
        },

        clear: () => {
            rmCache(KeyNotifications);
            set(emptyData);
        }
    };
}

const notifications = wrapper();

export {
    notifications
};
