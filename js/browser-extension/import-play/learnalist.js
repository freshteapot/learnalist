var el = document.querySelector('script[data-lal="1"]');
var extensionId = el.dataset.id;
var kind = el.dataset.kind;

setTimeout(function () {
    try {
        const dataElement = document.querySelector("#play-data");
        if (dataElement === null) {
            return
        }
        const aList = JSON.parse(dataElement.innerText);
        chrome.runtime.sendMessage(extensionId, {
            kind: kind,
            detail: aList,
            metadata: {
                kind: kind,
                extId: aList.uuid,
                refUrl: window.location.href,
                uuid: aList.uuid
            }
        });
    } catch (e) {
        chrome.runtime.sendMessage(extensionId, {
            kind: kind,
            detail: null
        });
    }
}, 0);
