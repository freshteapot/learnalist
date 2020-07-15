import { writable } from "svelte/store";
import { today, history, save } from "../plank/api.js";
import { get as cacheGet, save as cacheSave } from "../utils/storage.js";
import { copyObject } from "../utils/utils.js";
import { loggedIn } from "../store.js";
import dayjs from "dayjs";
import isToday from "dayjs/plugin/isToday";

dayjs.extend(isToday);


const StorageKeyPlankSettings = "plank.settings";
const StorageKeyPlankSavedItems = "plank.saved.items";

const emptyData = { today: {}, history: [] }

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

    allLists = response;
    const reduced = response.reduce(function (filtered, item) {
      filtered.push(...item.data);
      return filtered;
    }, []);

    data.history = reduced.reverse();

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

const loadToday = async () => {
  if (!loggedIn()) {
    data.today = {};
    set(data);
    return
  }

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

// Remove record
// Find which day the record is on and remove it
const deleteRecord = async (entry) => {
  const uuid = entry.uuid;
  const index = entry.listIndex;
  const found = allLists.find(aList => {
    return aList.uuid == uuid;
  });

  found.data.splice(index, 1);

  try {
    error.set('');
    loading.set(true);

    await save(found);

    if (dayjs(entry.beginningTime).isToday()) {
      await loadToday();
    }

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

  deleteRecord,

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
