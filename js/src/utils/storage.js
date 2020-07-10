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


export {
    get,
    save,
    rm
}
