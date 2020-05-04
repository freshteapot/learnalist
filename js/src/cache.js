const KeySettingsServer = "settings.server";
const KeySettingsInstallDefaults = "settings.install.defaults";
const KeyUserUuid = "app.user.uuid";
const KeyUserAuthentication = "app.user.authentication"
const KeyNotifications = "app.notifications";
const KeyEditorMyEditedLists = "my.edited.lists";
const KeyEditorMyLists = "my.lists";

function get(key, defaultResult) {
  if (!localStorage.hasOwnProperty(key)) {
    return defaultResult;
  }

  return JSON.parse(localStorage.getItem(key))
}

function save(key, data) {
  localStorage.setItem(key, JSON.stringify(data));
}

function rm(key) {
  localStorage.removeItem(key);
}

function clear() {
  localStorage.clear();
  save(KeySettingsInstallDefaults, true);
  const apiServer = document.querySelector('meta[name="api.server"]');
  if (apiServer) {
    save(KeySettingsServer, apiServer.content);
  } else {
    save(KeySettingsServer, "https://learnalist.net");
  }

  save(KeyEditorMyEditedLists, []);
  save(KeyEditorMyLists, []);
}

export {
  KeySettingsServer,
  KeySettingsInstallDefaults,
  KeyUserUuid,
  KeyUserAuthentication,
  KeyNotifications,
  get,
  save,
  rm,
  clear
};
