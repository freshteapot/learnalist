import { writable } from "svelte/store";
import { api } from "../../shared.js";
import { copyObject } from "../../utils/utils.js";

const emptyData = {};

let loaded = false;
let data = copyObject(emptyData);
const { subscribe, set, update } = writable(data);
const loading = writable(false);
const error = writable('');


const load = async (aList) => {
  set(aList);
  loaded = true;
}

const save = async (input) => {
  try {
    error.set('');
    loading.set(true);

    console.log(data === input);
    input.info.type = "v2";
    // TODO this is not being saved, I suspect due to openapi
    input.info.from = "https://quizlet.com/71954111/norwegian-flash-cards/";

    let aList = await api.addList(input);
    set(aList);

  } catch (e) {
    console.log(e);
    loading.set(false);
    error.set(`Error has been occurred. Details: ${e.message}`);
  }

}

const ImportPlayStore = () => ({
  subscribe,
  loading,
  error,
  load,
  save,
  loaded: () => loaded,
  getServer: () => api.getServer()
});

export default ImportPlayStore();
