(async () => {
  const config = await getConfig();
  const allowed = config.allowed;

  const extensionId = chrome.runtime.id;

  function loadScript(obj) {
    var s = document.createElement('script');
    s.src = chrome.runtime.getURL(obj.script);
    s.dataset.lal = "1";
    s.dataset.kind = obj.kind;
    s.dataset.id = extensionId;
    (document.head || document.documentElement).appendChild(s);
    s.onload = function () {
      s.remove();
    };
  }

  function whereAmI(location) {
    if (!allowed.map(el => el.origin).includes(location.origin)) {
      return null;
    }

    const found = allowed.find(el => el.origin == location.origin);
    return found;
  }


  chrome.runtime.onMessage.addListener(
    function (msg) {
      const allowed = ["lookup-login-info", "load-data"];

      if (!allowed.includes(msg.kind)) {
        console.log("kind not supported", msg.kind);
        return;
      }

      switch (msg.kind) {
        case "lookup-login-info":
          handleLogInToLearnalist(window.location);
          return;
        case "load-data":
          const kind = whereAmI(window.location);
          if (!kind) {
            chrome.runtime.sendMessage({ kind: "not-supported", location: window.location });
            return;
          }
          loadScript(kind);
          return;
      }
    }
  );

  async function handleLogInToLearnalist(location) {
    const config = await getConfig();
    const baseUrl = config.baseUrl;

    if (location.origin != baseUrl) {
      return;
    }

    // /logout.html redirects and seems to be to quick
    if (location.pathname == "/come-back-soon.html") {
      chrome.runtime.sendMessage({
        kind: 'learnalist-logout',
      });
      return;
    }

    const user = fromLocalStorage("app.user.uuid")
    const token = fromLocalStorage("app.user.authentication")
    if (user && token) {

      chrome.runtime.sendMessage({
        kind: 'learnalist-login-info',
        detail: {
          user: user,
          token: token,
        }
      });
    }
  }

  handleLogInToLearnalist(window.location);


  // Inform them we have logged in?
  // shared.notify("info", "You are not logged into the learnalist extension")
})()
