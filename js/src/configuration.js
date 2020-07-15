import { get, save, rm } from "./utils/storage.js";
import { getApiServer } from "./utils/setup.js";

const KeySettingsServer = "settings.server";
const KeySettingsInstallDefaults = "settings.install.defaults";
const KeyUserUuid = "app.user.uuid";
const KeyUserAuthentication = "app.user.authentication"
const KeyNotifications = "app.notifications";
const KeyEditorMyEditedLists = "my.edited.lists";
const KeyEditorMyLists = "my.lists";
const KeyLastScreen = "last.screen";
// Hmmm
const KeyListsByMe = "lists.by.me";

function clear() {
  console.log("clearing configuration");
  localStorage.clear();
  save(KeySettingsInstallDefaults, true);
  save(KeySettingsServer, getApiServer())
  save(KeyEditorMyEditedLists, []);
  save(KeyEditorMyLists, []);
}

const clearConfiguration = clear;
const saveConfiguration = save;
const removeConfiguration = rm;
const getConfiguration = get;

export {
  KeySettingsServer,
  KeySettingsInstallDefaults,
  KeyUserUuid,
  KeyUserAuthentication,
  KeyNotifications,
  KeyLastScreen,
  KeyListsByMe,
  KeyEditorMyEditedLists,
  KeyEditorMyLists,
  getConfiguration,
  saveConfiguration,
  removeConfiguration,
  clearConfiguration
};
