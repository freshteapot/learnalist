var el = document.querySelector('script[data-lal="1"]');
var extensionId = el.dataset.id;
var kind = el.dataset.kind;

setTimeout(function () {
    const extId = document.querySelector("#setID").value;
    chrome.runtime.sendMessage(extensionId, {
        kind: kind,
        detail: {
            title: document.querySelector("h1[itemprop=name]").innerHTML,
            listData: Cards
        },
        metadata: {
            kind: kind,
            extId: extId,
            refUrl: window.location.href
        }
    });
}, 0);
