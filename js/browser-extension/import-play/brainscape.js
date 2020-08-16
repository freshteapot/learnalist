var el = document.querySelector('script[data-lal="1"]');
var extensionId = el.dataset.id;
var kind = el.dataset.kind;

setTimeout(function () {
    const listData = Object.values(document.querySelectorAll("section.card")).map(el => {
        return {
            from: el.querySelector("h3.card-answer-text").innerText,
            to: el.querySelector("h2.card-question-text").innerText
        };
    });
    // Poor mans extraction of the title, ignoring the username
    const title = document.querySelector("title").innerText.split(" by ", 1)[0].trim();
    chrome.runtime.sendMessage(extensionId, {
        kind: kind,
        detail: {
            title: title,
            listData: listData
        }
    });
}, 0);
