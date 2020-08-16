<script>
  import Modal from "./spaced_repetition_modal.svelte";
  import { addEntry } from "./api.js";

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
  let showKey = "from";
  let show = false;

  function handleClose(event) {
    playElement.style.display = "none";
    push("/");
  }

  function edit(event) {
    // How did this work before?
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
    state = "edit";
    showKey = "from";
    show = false;
  }

  async function add(event) {
    const input = {
      show: data[showKey],
      data: data,
      settings: {
        show: showKey
      },
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
  <table class="w-100" cellspacing="0">
    <thead>
      <tr>
        <th class="fw6 bb b--black-20 pb3 tl">From</th>
        <th class="fw6 bb b--black-20 pb3 tl">To</th>
      </tr>
    </thead>
    <tbody class="lh-copy">
      {#each aList.data as item, index}
        <tr data-index={index} on:click={edit}>
          <td class="pv3 pr3 bb b--black-20">{item.from}</td>
          <td class="pv3 pr3 bb b--black-20">{item.to}</td>
        </tr>
      {/each}
    </tbody>
  </table>
</div>

<Modal {show} {state} on:add={add} on:close={close}>
  {#if state === 'edit'}
    <p>
      <span>Which to show?</span>
    </p>
    <p>
      <input type="radio" bind:group={showKey} value={'from'} />
      from
    </p>
    <p>
      <input type="radio" bind:group={showKey} value={'to'} />
      to
    </p>
    <pre>{JSON.stringify(data, '', 2)}</pre>
  {/if}

  {#if state === 'feedback'}
    <p>Already in the system</p>
    <p>You will be reminded on {data.settings.when_next}</p>
  {/if}
</Modal>
