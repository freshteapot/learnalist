import { writable } from "svelte/store";
import { getVersion } from "../../api.js";


const { subscribe, set, update } = writable({
  "gitHash": "na",
  "gitDate": "na",
  "version": "na",
  "url": "https://github.com/freshteapot/learnalist-api"
}
);

const loading = writable(false);
const error = writable('');

const VersionStore = () => ({
  subscribe,
  set,
  loading,
  error,
  async get(query) {
    console.log("Not here");
    try {
      error.set('');
      loading.set(true);
      const response = await getVersion();
      loading.set(false);
      set(response);
      return response;
    } catch (e) {
      loading.set(false);
      set({
        "gitHash": "na",
        "gitDate": "na",
        "version": "na",
        "url": "https://github.com/freshteapot/learnalist-api"
      });
      error.set(`Error has been occurred. Details: ${e.message}`);
    }
  }
});

export default VersionStore();
