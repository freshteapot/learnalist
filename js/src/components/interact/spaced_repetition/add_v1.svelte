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
  let state = "edit";
  let show = false;

  function edit(event) {
    data = event.detail.data;
    show = true;
  }

  function close() {
    data = null;
    show = false;
    state = "edit";
  }

  async function add(event) {
    const input = {
      show: data,
      data: data,
      kind: aList.info.type
    };
    const response = await addEntry(input);

    switch (response.status) {
      case 201:
        close();
        break;
      case 200:
        state = "feedback";
        data = response.body;
        break;
      default:
        console.log("failed to add for spaced learning");
        console.log(response);
        break;
    }
  }
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
  {state}
  on:add={add}
  on:edit={edit}
  on:close={close}>

  {#if state === 'edit'}
    <pre>{JSON.stringify(data, '', 2)}</pre>
  {/if}

  {#if state === 'feedback'}
    <p>Already in the system</p>
    <p>You will be reminded on {data.settings.when_next}</p>
  {/if}
</Modal>
