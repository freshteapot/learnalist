import { writable } from "svelte/store";
import { api } from "../../shared.js";
import { copyObject } from "../../utils/utils.js";

const emptyData = {};

let loaded = false;
let aListData = copyObject(emptyData);
const aList = writable(aListData);

const load = async (input) => {
  aListData = input;
  aList.set(aListData);
  loaded = true;
}

const save = async () => {
  try {
    const input = aListData;
    input.info.type = "v2";
    input.info.from.ext_uuid = input.info.from.ext_uuid.toString();
    aListData = await api.addList(input);
    aList.set(aListData);

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
