import { get, save, rm } from "./utils/storage.js";
import { getApiServer } from "./utils/setup.js";

const KeySettingsServer = "settings.server";
const KeySettingsInstallDefaults = "settings.install.defaults";
const KeyUserUuid = "app.user.uuid";
const KeyUserAuthentication = "app.user.authentication"
const KeyNotifications = "app.notifications";
const KeyEditorMyEditedLists = "my.edited.lists";
const KeyEditorMyLists = "my.lists";

function clear() {
  localStorage.clear();
  save(KeySettingsInstallDefaults, true);
  save(KeySettingsServer, getApiServer())
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
