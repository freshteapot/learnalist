import { writable } from 'svelte/store';
import cache from './cache.js';

const data = {
    level: "",
    message: "",
}

const emptyData = JSON.parse(JSON.stringify(data));
let liveData = JSON.parse(JSON.stringify(data));


const storedData = cache.get(cache.KeyNotifications, null);

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
                cache.save(cache.KeyNotifications, notification)
                return notification;
            });
        },

        clear: () => {
            cache.rm(cache.KeyNotifications);
            set(emptyData);
        }
    };
}

const notifications = wrapper();

export {
    notifications
};
