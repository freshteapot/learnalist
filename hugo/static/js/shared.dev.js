var superstore = (function (exports) {
    'use strict';

    function noop() { }
    function safe_not_equal(a, b) {
        return a != a ? b == b : a !== b || ((a && typeof a === 'object') || typeof a === 'function');
    }

    const subscriber_queue = [];
    /**
     * Create a `Writable` store that allows both updating and reading by subscription.
     * @param {*=}value initial value
     * @param {StartStopNotifier=}start start and stop notifications for subscriptions
     */
    function writable(value, start = noop) {
        let stop;
        const subscribers = [];
        function set(new_value) {
            if (safe_not_equal(value, new_value)) {
                value = new_value;
                if (stop) { // store is ready
                    const run_queue = !subscriber_queue.length;
                    for (let i = 0; i < subscribers.length; i += 1) {
                        const s = subscribers[i];
                        s[1]();
                        subscriber_queue.push(s, value);
                    }
                    if (run_queue) {
                        for (let i = 0; i < subscriber_queue.length; i += 2) {
                            subscriber_queue[i][0](subscriber_queue[i + 1]);
                        }
                        subscriber_queue.length = 0;
                    }
                }
            }
        }
        function update(fn) {
            set(fn(value));
        }
        function subscribe(run, invalidate = noop) {
            const subscriber = [run, invalidate];
            subscribers.push(subscriber);
            if (subscribers.length === 1) {
                stop = start(set) || noop;
            }
            run(value);
            return () => {
                const index = subscribers.indexOf(subscriber);
                if (index !== -1) {
                    subscribers.splice(index, 1);
                }
                if (subscribers.length === 0) {
                    stop();
                    stop = null;
                }
            };
        }
        return { set, update, subscribe };
    }

    const KeyUserAuthentication = "app.user.authentication";
    const KeyNotifications = "app.notifications";

    function get(key, defaultResult) {
      if (!localStorage.hasOwnProperty(key)) {
        return defaultResult;
      }

      return JSON.parse(localStorage.getItem(key))
    }

    function save(key, data) {
      localStorage.setItem(key, JSON.stringify(data));
    }

    function rm(key) {
      localStorage.removeItem(key);
    }

    const data = {
        level: "",
        message: "",
    };

    const emptyData = JSON.parse(JSON.stringify(data));
    let liveData = JSON.parse(JSON.stringify(data));


    const storedData = get(KeyNotifications, null);

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
                    save(KeyNotifications, notification);
                    return notification;
                });
            },

            clear: () => {
                rm(KeyNotifications);
                set(emptyData);
            }
        };
    }

    const notifications = wrapper();

    // Link any component to be able to notify the banner component
    const notify = (level, message) => {
        notifications.add(level, message);
    };

    const loggedIn = () => {
        return localStorage.hasOwnProperty(KeyUserAuthentication);
    };

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
    };

    exports.loggedIn = loggedIn;
    exports.login = login;
    exports.notifications = notifications;
    exports.notify = notify;

    return exports;

}({}));
//# sourceMappingURL=shared.dev.js.map
