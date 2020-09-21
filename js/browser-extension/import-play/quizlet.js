var el = document.querySelector('script[data-lal="1"]');
var extensionId = el.dataset.id;
var kind = el.dataset.kind;

setTimeout(function () {
    try {
        chrome.runtime.sendMessage(extensionId, {
            kind: kind,
            detail: {
                title: window.Quizlet.setPageData.set.title,
                listData: window.Quizlet
            },
            metadata: {
                kind: kind,
                ext_uuid: window.Quizlet.setPageData.set.id,
                ref_url: window.location.href
            }
        });
    } catch (e) {
        chrome.runtime.sendMessage(extensionId, {
            kind: kind,
            detail: null
        });
    }
}, 0);

