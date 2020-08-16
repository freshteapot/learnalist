var el = document.querySelector('script[data-lal="1"]');
var extensionId = el.dataset.id;
var kind = el.dataset.kind;

setTimeout(function () {
    chrome.runtime.sendMessage(extensionId, {
        kind: kind,
        detail: {
            title: document.querySelector("h1[itemprop=name]").innerHTML,
            listData: Cards
        }
    });
}, 0);
