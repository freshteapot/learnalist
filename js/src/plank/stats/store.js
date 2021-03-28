import { writable } from "svelte/store";
import { history } from "../api.js";

import { copyObject } from "../../utils/utils.js";
import { loggedIn } from "../../shared.js";

function Create() {
  const emptyData = { history: [] }
  let data = copyObject(emptyData);
  const { subscribe, set, update } = writable(data);

  const loadHistory = async () => {
    if (!loggedIn()) {
      data.history = [];
      set(data);
      return
    }

    try {
      const response = await history();
      data.history = response;
      set(data);
    } catch (e) {
      console.log(e);
      data = copyObject(emptyData);
      set(data);
    }
  }

  return {
    subscribe,
    loadHistory,
  };


}

//export default PlankStore();
export const planks = Create();
