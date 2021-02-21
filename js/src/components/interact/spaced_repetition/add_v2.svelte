<svelte:options tag={null} accessors={true} />

<script>
  import { push } from "svelte-spa-router";
  import { onMount } from "svelte";
  import OvertimeActive from "./overtime_active.svelte";
  import Modal from "./spaced_repetition_modal.svelte";
  import {
    addEntry,
    addListToOvertime,
    overtimeIsActive,
  } from "../../../spaced_repetition/api.js";
  import { loggedIn, notify } from "../../../shared.js";
  import LoginModal from "../../../components/login_modal.svelte";
  import { KeyUserUuid, getConfiguration } from "../../../configuration";

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

  let overtimeActive = false;
  let showAddingOvertime = false;
  let userUuid = "";

  const loginNagMessageDefault =
    "You need to be logged in so we can personalise your learning experience.";
  let loginNagMessage = loginNagMessageDefault;
  let loginNagClosed = true;
  let listIsEmpty = aList.data.length === 0;

  onMount(async () => {
    userUuid = getConfiguration(KeyUserUuid);

    if (loggedIn()) {
      overtimeActive = await overtimeIsActive(aList.uuid);
    }
  });

  function handleClose(event) {
    playElement.style.display = "none";
    push("/");
  }

  function edit(event) {
    if (!loggedIn()) {
      loginNagClosed = false;
      return;
    }

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
        show: showKey,
      },
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
    const input = {
      alist_uuid: aList.uuid,
      user_uuid: userUuid,
      settings: {
        show: showKey,
      },
    };
    const added = await addListToOvertime(input);
    // TODO maybe visualise it failed
    overtimeActive = added;
    closeOvertime();
  }

  function addingOvertime() {
    if (listIsEmpty) {
      notify("error", "No items to add", false);
      return;
    }

    if (!loggedIn()) {
      loginNagClosed = false;
      return;
    }

    data = aList.data[0];
    showAddingOvertime = true;
  }

  function closeOvertime() {
    showAddingOvertime = false;
    close();
  }
</script>

{#if overtimeActive}
  <div class="flex flex-column">
    <div class=" w-100 pa3 mr2">
      <header>
        <button class="br3" on:click={handleClose}>Close</button>
        <h1 class="f2 measure" title="Spaced Repetition">🧠 + 💪</h1>
        <OvertimeActive alistUuid={aList.uuid} {userUuid} bind:overtimeActive />
      </header>
    </div>
  </div>
{/if}

{#if !overtimeActive}
  <div class="flex flex-column">
    <div class=" w-100 pa3 mr2">
      <header>
        <h1 class="f2 measure" title="Spaced Repetition">🧠 + 💪</h1>
        <button class="br3" on:click={handleClose}>Close</button>
        <p>
          Click on the row you want to add or <button
            class="br3"
            on:click|preventDefault={addingOvertime}>add all overtime</button
          >
        </p>
      </header>
    </div>

    <div id="list-data" class=" w-100 pa3 mr2">
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
  </div>

  {#if showAddingOvertime}
    <Modal
      show="true"
      state="edit"
      on:add={addOvertime}
      on:close={closeOvertime}
    >
      <p>
        <span>Which to show?</span>
      </p>
      <p>
        <input type="radio" bind:group={showKey} value={"from"} />
        from
      </p>
      <p>
        <input type="radio" bind:group={showKey} value={"to"} />
        to
      </p>
      <pre>{JSON.stringify(data, '', 2)}</pre>
    </Modal>
  {/if}

  <Modal {show} {state} on:add={add} on:close={close}>
    {#if state === "edit"}
      <p>
        <span>Which to show?</span>
      </p>
      <p>
        <input type="radio" bind:group={showKey} value={"from"} />
        from
      </p>
      <p>
        <input type="radio" bind:group={showKey} value={"to"} />
        to
      </p>
      <pre>{JSON.stringify(data, '', 2)}</pre>
    {/if}

    {#if state === "feedback"}
      <p>Already in the system</p>
      <p>You will be reminded on {data.settings.when_next}</p>
    {/if}
  </Modal>
{/if}

{#if !loggedIn() && !loginNagClosed}
  <LoginModal on:close={(e) => (loginNagClosed = true)}>
    <p>{loginNagMessage}</p>
  </LoginModal>
{/if}

<style>
  @import "../../../../all.css";
</style>
