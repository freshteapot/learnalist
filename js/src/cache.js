const KeySettingsServer = "settings.server";
const KeySettingsInstallDefaults = "settings.install.defaults";
const KeyAuthentication = "settings.authentication";
const KeyUserUuid = "app.user.uuid";
const KeyNotifications = "app.notifications";


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
  save(KeySettingsServer, "https://learnalist.net");
}

export default {
  KeyAuthentication,
  KeySettingsServer,
  KeySettingsInstallDefaults,
  KeyUserUuid,
  KeyNotifications,
  get,
  save,

  rm,
  clear
};
