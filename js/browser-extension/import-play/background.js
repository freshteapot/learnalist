(async () => {
  function handleActivated(activeInfo) {
    console.log("Tab " + activeInfo.tabId +
      " was activated", activeInfo);

    chrome.tabs.get(activeInfo.tabId, function (tab) {
      console.log('New active tab: ' + tab.id, tab);
    });
  }

  chrome.tabs.onActivated.addListener(handleActivated);


  async function findLearnalist(windows) {
    const config = await getConfig();
    const baseUrl = config.baseUrl;

    const learnalistTab = windows[0].tabs.find(tab => {
      return tab.url.includes(baseUrl);
    });

    if (!learnalistTab) {
      return;
    }

    const url = new URL(learnalistTab.url);

    if (url.origin != baseUrl) {
      return;
    }

    chrome.tabs.sendMessage(learnalistTab.id, { kind: "lookup-login-info" });
  }


  chrome.windows.getAll({ populate: true }, findLearnalist);

  chrome.runtime.onMessage.addListener(async (msg, sender) => {

    const config = await getConfig();
    const baseUrl = config.baseUrl;

    if (msg.kind == "learnalist-login-info") {
      if (sender.id != chrome.runtime.id) {
        return;
      }

      if (sender.origin != baseUrl) {
        return;
      }

      toLocalStorage("app.user.uuid", msg.detail.user);
      toLocalStorage("app.user.authentication", msg.detail.token);
      return;
    }

    if (msg.kind == "learnalist-logout") {
      if (sender.id != chrome.runtime.id) {
        return;
      }

      if (sender.origin != baseUrl) {
        return;
      }
      localStorage.removeItem("app.user.uuid");
      localStorage.removeItem("app.user.authentication");
      return;
    }

    if (msg.kind == "lookup-login-info") {
      chrome.windows.getAll({ populate: true }, findLearnalist);
      return;
    }
  });

  addSpacedRepetitionMenu();
})()
