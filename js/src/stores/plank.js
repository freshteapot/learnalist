import { writable } from "svelte/store";
import { today, history, save } from "../plank/api.js";
import { get as cacheGet, save as cacheSave } from "../utils/storage.js";
import { copyObject } from "../utils/utils.js";

const StorageKeyPlankSettings = "plank.settings";
const emptyData = { today: {}, history: [] }

let data = copyObject(emptyData);
const { subscribe, set, update } = writable(data);


const loading = writable(false);
const error = writable('');

const loadHistory = async () => {
  try {
    error.set('');
    loading.set(true);
    const response = await history();
    loading.set(false);
    data.history = response.reverse();
    set(data);
  } catch (e) {
    console.log(e);
    loading.set(false);
    data = copyObject(emptyData);
    set(data);
    error.set(`Error has been occurred. Details: ${e.message}`);
  }
}

const loadToday = async () => {
  try {
    error.set('');
    loading.set(true);
    const response = await today();
    loading.set(false);
    data.today = response;
    set(data);
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
  today() {
    return copyObject(data.today);
  },

  history() {
    return copyObject(data.history);
  },

  async record(aList) {
    try {
      error.set('');
      loading.set(true);

      await save(aList);
      await loadToday();
      await loadHistory();
    } catch (e) {
      console.log(e);
      loading.set(false);
      data = copyObject(emptyData);
      set(data);
      error.set(`Error has been occurred. Details: ${e.message}`);
    }
  },

  today: loadToday,
  history: loadHistory,

  settings(input) {
    if (!input) {
      return cacheGet(StorageKeyPlankSettings, { showIntervals: false, intervalTime: 15 });
    }
    cacheSave(StorageKeyPlankSettings, { showIntervals: input.showIntervals, intervalTime: input.intervalTime });
  }
});

export default PlankStore();
