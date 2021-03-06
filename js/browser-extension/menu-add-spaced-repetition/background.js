function addSpacedRepetitionMenu() {
  chrome.contextMenus.create({
    id: "spaced-repetition-add",
    title: "Add to 🧠 + 💪",
    contexts: ["selection"],
    onclick: (info) => {
      chrome.tabs.create({
        url: `https://learnalist.net/spaced-repetition.html#/add?c=${info.selectionText}`
      });
    }
  });
}

addSpacedRepetitionMenu();
