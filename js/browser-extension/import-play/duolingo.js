var el = document.querySelector('script[data-lal="1"]');
var extensionId = el.dataset.id;
var kind = el.dataset.kind;

setTimeout(function () {
    try {
        // There is an id in the json response, but not sure how to find it
        const extId = window.location.pathname;
        //const title = document.querySelector('[data-test="back-arrow"] ~ div').innerText;

        const elTable = Object.values(
            document.querySelectorAll("table")
        ).find(el => {
            return el.querySelector("thead tr th:first-child").innerText === "Vocabulary";
        });

        const title = elTable.parentElement.previousElementSibling.innerText

        const listData = Object.values(
            elTable.querySelectorAll("tbody tr")
        ).map(el => {
            return {
                from: el.querySelectorAll("td")[0].innerText,
                to: el.querySelectorAll("td")[1].innerText
            };
        });

        chrome.runtime.sendMessage(extensionId, {
            kind: kind,
            detail: {
                title: title,
                listData: listData
            },
            metadata: {
                kind: kind,
                ext_uuid: extId,
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
