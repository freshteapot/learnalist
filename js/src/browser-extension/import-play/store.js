import { writable } from "svelte/store";
import { api } from "../../shared.js";
import { copyObject } from "../../utils/utils.js";

const emptyData = {};

let loaded = false;
let data = copyObject(emptyData);
const aList = writable(data);

const load = async (input) => {
  aList.set(input);
  loaded = true;
}

const save = async () => {
  try {
    const input = aList.get();
    input.info.type = "v2";
    input.info.from.ext_uuid = input.info.from.ext_uuid.toString();
    let response = await api.addList(input);
    aList.set(response);

  } catch (e) {
    throw new Error(e.message);
  }

}
const ImportPlayStore = () => ({
  load,
  save,
  loaded: () => loaded,
  getServer: () => api.getServer(),
  aList
});

export default ImportPlayStore();
