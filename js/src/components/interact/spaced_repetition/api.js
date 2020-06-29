
import { getServer, getHeaders } from '../../../api.js';

async function getNext() {
    const url = getServer() + "/api/v1/spaced-repetition/next";
    let headers;
    try {
        headers = getHeaders();
    } catch (error) {
        let response = {
            status: 403,
            body: {}
        };
        return response;
    }

    const res = await fetch(url, {
        headers: headers
    });

    let response = {
        status: res.status,
        body: {}
    };

    if (res.status == 200) {
        response.body = await res.json();
        return response;
    }

    return response;
}

async function viewed(uuid) {
    const input = {
        uuid: uuid,
        action: "incr"
    }

    const url = getServer() + "/api/v1/spaced-repetition/viewed";
    const res = await fetch(url, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(input)
    });

    return res.status;
}

async function addEntry(input) {
    const url = getServer() + "/api/v1/spaced-repetition/";
    const res = await fetch(url, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(input)
    });

    let response = {
        status: res.status,
        body: await res.json(),
    };

    return response;
}

export {
    getNext,
    viewed,
    addEntry,
};
