<script>
  import Modal from "./spaced_repetition_modal.svelte";
  import { addEntry } from "../../../spaced_repetition/api.js";

  import { push } from "svelte-spa-router";
  import { tap } from "@sveltejs/gestures";

  // {DomElement}
  export let playElement;
  // {DomElement}
  export let listElement;
  // learnalist aList object
  export let aList;

  playElement.style.display = "";
  listElement.style.display = "none";

  let data;
  let state = "edit";
  let show = false;

  function handleClose(event) {
    playElement.style.display = "none";
    push("/");
  }

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

<header>
  <button class="br3" on:click={handleClose}>Close</button>
  <h1 class="f2 measure" title="Spaced Repetition">🧠 + 💪</h1>
  <h3>Click on the row you want to add</h3>
</header>

<div id="list-data">
  <ul class="lh-copy ph0 list">
    {#each aList.data as item, index}
      <li class="pv3 pr3 bb b--black-20" data-index={index}>{item}</li>
    {/each}
  </ul>
</div>

<Modal {show} {state} on:add={add} on:close={close}>

  {#if state === 'edit'}
    <pre>{JSON.stringify(data, '', 2)}</pre>
  {/if}

  {#if state === 'feedback'}
    <p>Already in the system</p>
    <p>You will be reminded on {data.settings.when_next}</p>
  {/if}
</Modal>
