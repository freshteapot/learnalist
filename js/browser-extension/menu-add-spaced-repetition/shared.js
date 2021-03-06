async function openSpacedRepetition(info, tab) {
    const config = await getConfig();

    chrome.tabs.create({
        url: `${config.baseUrl}/spaced-repetition.html#/add?c=${info.selectionText}`
    });
}

function addSpacedRepetitionMenu() {
    chrome.contextMenus.create({
        id: "spaced-repetition-add",
        title: "Add to 🧠 + 💪",
        contexts: ["selection"],
        onclick: openSpacedRepetition
    });
}
