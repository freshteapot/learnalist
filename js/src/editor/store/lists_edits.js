import cache from "../lib/cache.js";
import { writable } from 'svelte/store';

const current = cache.get(cache.keys["my.edited.lists"]);
const {subscribe, set, update} = writable(current);

const ListsEditsStore = () => ({
    subscribe,
    set,

    find(uuid) {
      let found;
      update(edits => {
        found = edits.find(aList => {
            return aList.uuid === uuid
        })
        return edits;
      });
      return found;
    },

    add(aList) {
      update(edits => {
        const found = edits.some(item => item.uuid === aList.uuid);
        if (!found) {
          edits.push(aList);
        }
        return edits;
      });
    },

    update(aList) {
      update(edits => {
        const updated = edits.map(item => {
          if (item.uuid === aList.uuid) {
            item = aList;
          }
          return item;
        });
        cache.save(cache.keys["my.edited.lists"], updated);
        return updated;
      });
    },

    remove(uuid) {
      update(edits => {
        const found = edits.filter(aList => aList.uuid !== uuid);
        cache.save(cache.keys["my.edited.lists"], found);
        return found;
      });
    }
});

export default ListsEditsStore();
