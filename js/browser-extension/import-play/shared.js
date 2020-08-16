// ---
async function openSpacedRepetition(info, tab) {
    const config = await getConfig();

    chrome.tabs.create({
        url: `${config.baseUrl}/spaced-repetition.html#/add?c=${info.selectionText}`
    });
}

function addSpacedRepetitionMenu() {
    console.log("adding spaced-repetition-add");
    chrome.contextMenus.create({
        id: "spaced-repetition-add",
        title: "Add to 🧠 + 💪",
        contexts: ["selection"],
        onclick: openSpacedRepetition
    });
}

// ---


function fromLocalStorage(key) {
    try {
        return JSON.parse(localStorage.getItem(key));
    } catch (e) {
        return null;
    }
}

function toLocalStorage(key, value) {
    localStorage.setItem(key, JSON.stringify(value));
}

async function getConfig() {
    const config = await fetch(chrome.runtime.getURL('config.json'))
        .then((response) => response.json());

    if (!localStorage.hasOwnProperty("settings.server")) {
        toLocalStorage("settings.server", config.baseUrl);
    }

    const baseUrl = fromLocalStorage("settings.server");
    config.baseUrl = baseUrl;
    return config;
}
