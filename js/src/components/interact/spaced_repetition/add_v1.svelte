<svelte:options tag={null} accessors={true} />

<script>
  import Modal from "./spaced_repetition_modal.svelte";
  import OvertimeActive from "./overtime_active.svelte";
  import { addEntry } from "../../../spaced_repetition/api.js";

  import { push } from "svelte-spa-router";
  import { onMount } from "svelte";
  import { api } from "../../../shared";
  import { KeyUserUuid, getConfiguration } from "../../../configuration";
  import { prevent_default } from "svelte/internal";

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
  // Check to see if list is already added over time
  let overtimeActive = false;
  let userUuid = "";

  onMount(async () => {
    userUuid = getConfiguration(KeyUserUuid);
    overtimeActive = await api.spacedRepetitionOvertimeIsActive(aList.uuid);
  });

  function handleClose(event) {
    playElement.style.display = "none";
    push("/");
  }

  function edit(event) {
    const index = event.target
      .closest("[data-index]")
      .getAttribute("data-index");

    if (!index) {
      return;
    }

    data = aList.data[index];
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
      kind: aList.info.type,
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

  async function addOvertime() {
    console.log("Add overtime");
    const input = {
      alist_uuid: aList.uuid,
      user_uuid: userUuid,
    };
    const added = await api.spacedRepetitionAddListToOvertime(input);
    // TODO maybe visualise it failed
    overtimeActive = added;
  }
</script>

{#if overtimeActive}
  <header>
    <button class="br3" on:click={handleClose}>Close</button>
    <h1 class="f2 measure" title="Spaced Repetition">🧠 + 💪</h1>
    <OvertimeActive alistUuid={aList.uuid} {userUuid} bind:overtimeActive />
  </header>
{/if}

{#if !overtimeActive}
  <header>
    <button class="br3" on:click={handleClose}>Close</button>
    <h1 class="f2 measure" title="Spaced Repetition">🧠 + 💪</h1>
    <p>
      Click on the row you want to add or <a
        href="#"
        class="link underline"
        on:click|once|preventDefault={addOvertime}>add all overtime</a
      >
    </p>
  </header>

  <div id="list-data">
    <ul class="lh-copy ph0 list">
      {#each aList.data as item, index}
        <li class="pv3 pr3 bb b--black-20" data-index={index} on:click={edit}>
          {item}
        </li>
      {/each}
    </ul>
  </div>

  <Modal {show} {state} on:add={add} on:close={close}>
    {#if state === "edit"}
      <pre>{JSON.stringify(data, '', 2)}</pre>
    {/if}

    {#if state === "feedback"}
      <p>Already in the system</p>
      <p>You will be reminded on {data.settings.when_next}</p>
    {/if}
  </Modal>
{/if}

<style>
  @import "../../../../all.css";
</style>
