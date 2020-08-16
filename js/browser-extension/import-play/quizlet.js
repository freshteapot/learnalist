var el = document.querySelector('script[data-lal="1"]');
var extensionId = el.dataset.id;
var kind = el.dataset.kind;

setTimeout(function () {
    chrome.runtime.sendMessage(extensionId, {
        kind: kind,
        detail: {
            title: window.Quizlet.setPageData.set.title,
            listData: window.Quizlet
        },
        metadata: {
            kind: kind,
            extId: window.Quizlet.setPageData.set.id,
            refUrl: window.location.href
        }
    });
}, 0);
