import { get, writable } from 'svelte/store';
import { getListsByMe } from "../api2.js";
import {
  getConfiguration,
  saveConfiguration,
  KeyListsByMe
} from "../configuration.js";


const current = getConfiguration(KeyListsByMe, []);
const { subscribe, set, update } = writable(current);
const loading = writable(false);
const error = writable('');

const ListsByMeStore = () => ({
  subscribe,
  set,
  loading,
  error,
  async get() {
    let data = [];
    try {
      data = getConfiguration(KeyListsByMe, []);
      set(data);
      error.set('');
      if (data.length === 0) {
        loading.set(true);
      }

      data = await getListsByMe();
      loading.set(false);
      saveConfiguration(KeyListsByMe, data);
      set(data);
      return data;
    } catch (e) {
      loading.set(false);
      data = getConfiguration(KeyListsByMe, []);
      set(data);
      error.set(`Error has been occurred. Details: ${e.message}`);
    }
  },

  find(uuid) {
    return get(this).find(aList => {
      return aList.uuid === uuid
    })
  },

  add(aList) {
    update(myLists => {
      myLists.push(aList);
      saveConfiguration(KeyListsByMe, myLists);
      return myLists;
    });
  },

  update(aList) {
    update(myLists => {
      const updated = myLists.map(item => {
        if (item.uuid === aList.uuid) {
          item = aList;
        }
        return item;
      });
      saveConfiguration(KeyListsByMe, updated);
      return updated;
    });
  },

  remove(uuid) {
    update(myLists => {
      const found = myLists.filter(aList => aList.uuid !== uuid);
      saveConfiguration(KeyListsByMe, found);
      return found;
    });
  }
});

export default ListsByMeStore();
