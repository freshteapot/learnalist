<script>
  import Modal from "./spaced_repetition_modal.svelte";
  import { addEntry } from "./api.js";

  import { push } from "svelte-spa-router";
  import { tap } from "@sveltejs/gestures";

  // {DomElement}
  export let listDataElement;
  // {DomElement}
  export let playElement;
  // {DomElement}
  export let listTitleElement;
  // learnalist aList object
  export let aList;

  let data;
  let show = false;
  function edit(event) {
    data = event.detail.data;
    show = true;
  }

  function close() {
    data = null;
    show = false;
  }

  async function add(event) {
    const input = {
      show: data,
      data: data,
      kind: aList.info.type
    };
    const response = await addEntry(input);

    if (response.status !== 200) {
      console.log("failed to add for spaced learning");
      console.log(response);
      return;
    }

    close();
  }

  // How to handle when not logged in
</script>

<style>
  @import "../../../../all.css";
</style>

<svelte:options tag={null} accessors={true} />

<Modal
  {aList}
  {listDataElement}
  {playElement}
  {listTitleElement}
  {show}
  on:add={add}
  on:edit={edit}
  on:close={close}>
  <pre>{JSON.stringify(data, '', 2)}</pre>
</Modal>
