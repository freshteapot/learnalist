import { writable } from "svelte/store";
import { history, saveEntry, deleteEntry } from "./api.js";
import { get as cacheGet, save as cacheSave } from "../utils/storage.js";
import { copyObject } from "../utils/utils.js";
import { loggedIn } from "../shared.js";
import dayjs from "dayjs";
import isToday from "dayjs/plugin/isToday";

dayjs.extend(isToday);

const StorageKeyPlankSettings = "plank.settings";
const StorageKeyPlankSavedItems = "plank.saved.items";

const emptyData = { history: [] }

let data = copyObject(emptyData);
const { subscribe, set, update } = writable(data);

// Contains all the lists with the labels plank + plank+YYMMDD
let allLists = [];
const loading = writable(false);
const error = writable('');

const loadHistory = async () => {
  if (!loggedIn()) {
    const tempHistory = cacheGet(StorageKeyPlankSavedItems, []);
    data.history = tempHistory.reverse();
    set(data);
    return
  }

  try {
    error.set('');
    loading.set(true);
    const response = await history();
    loading.set(false);
    data.history = response;
    set(data);
  } catch (e) {
    console.log(e);
    loading.set(false);
    allLists = [];
    data = copyObject(emptyData);
    set(data);
    error.set(`Error has been occurred. Details: ${e.message}`);
  }
}

// Remove record
// Find which day the record is on and remove it
const deleteRecord = async (entry) => {
  try {
    error.set('');
    loading.set(true);
    await deleteEntry(entry.uuid);
    await loadHistory();
  } catch (e) {
    console.log(e);
    loading.set(false);
    data = copyObject(emptyData);
    set(data);
    error.set(`Error has been occurred. Details: ${e.message}`);
  }
}

// If entry is not set we try
const record = async (entry) => {
  // TODO this will be greatly simplified
  if (entry) {
    // First we save to the temporary area.
    let items = cacheGet(StorageKeyPlankSavedItems, []);
    items.push(entry);
    cacheSave(StorageKeyPlankSavedItems, items);
  }

  if (!loggedIn()) {
    // Even when not logged in we are building the fake data structures
    await loadHistory();
    return
  }

  const items = cacheGet(StorageKeyPlankSavedItems, []);
  if (items.length == 0) {
    return;
  }

  try {
    error.set('');
    loading.set(true);

    await Promise.all(items.map(item => saveEntry(item)))
    cacheSave(StorageKeyPlankSavedItems, []);

    await loadHistory();
  } catch (e) {
    console.log(e);
    loading.set(false);
    data = copyObject(emptyData);
    set(data);
    error.set(`Error has been occurred. Details: ${e.message}`);
  }
}

const PlankStore = () => ({
  subscribe,
  loading,
  error,

  history() {
    return copyObject(data.history);
  },

  record,
  deleteRecord,
  history: loadHistory,

  settings(input) {
    if (!input) {
      return cacheGet(StorageKeyPlankSettings, { showIntervals: false, intervalTime: 15 });
    }
    cacheSave(StorageKeyPlankSettings, { showIntervals: input.showIntervals, intervalTime: input.intervalTime });
  }
});

export default PlankStore();
