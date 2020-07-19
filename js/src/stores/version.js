import { writable } from "svelte/store";
import { api } from "../shared.js";

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
  async get() {
    try {
      error.set('');
      loading.set(true);
      const response = await api.getServerVersion();
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
