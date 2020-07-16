import { writable } from 'svelte/store';
import {
    removeConfiguration,
    getConfiguration,
    saveConfiguration,
    KeyNotifications
} from './configuration.js';
import { copyObject } from './utils/utils.js';

const data = {
    level: "",
    message: "",
    sticky: false,
}

const emptyData = JSON.parse(JSON.stringify(data));
let liveData = JSON.parse(JSON.stringify(data));


const storedData = getConfiguration(KeyNotifications, null);

if (storedData !== null) {
    liveData = storedData;
}

const { subscribe, update, set } = writable(liveData);

function wrapper() {
    return {
        subscribe,

        add: (level, message, sticky) => {
            if (sticky == undefined) {
                sticky = false;
            }

            update(notification => {
                notification.level = level;
                notification.message = message;
                notification.sticky = sticky;
                saveConfiguration(KeyNotifications, notification)
                return notification;
            });
        },

        clear: () => {
            removeConfiguration(KeyNotifications);
            set(copyObject(emptyData));
        }
    };
}

const notifications = wrapper();

export {
    notifications
};
