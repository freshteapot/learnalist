import { get, writable } from 'svelte/store';
import { getListsByMe } from '../lib/api.js';
import cache from '../lib/cache.js';

const current = cache.get(cache.keys["lists.by.me"]);
const { subscribe, set, update } = writable(current);
const loading = writable(false);
const error = writable('');

const ListsByMeStore = () => ({
  subscribe,
  set,
  loading,
  error,
  async get() {
    let key = cache.keys['lists.by.me'];
    let data = [];
    try {
      data = cache.get(key, data);
      set(data);
      error.set('');
      if (data.length === 0) {
        loading.set(true);
      }

      const response = await getListsByMe();
      loading.set(false);
      cache.save(key, response);
      set(response);
      return response;
    } catch (e) {
      loading.set(false);
      data = cache.get(key, data);
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
      cache.save(cache.keys["lists.by.me"], myLists);
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
      cache.save(cache.keys["lists.by.me"], updated);
      return updated;
    });
  },

  remove(uuid) {
    update(myLists => {
      const found = myLists.filter(aList => aList.uuid !== uuid);
      cache.save(cache.keys["lists.by.me"], found);
      return found;
    });
  }
});

export default ListsByMeStore();
