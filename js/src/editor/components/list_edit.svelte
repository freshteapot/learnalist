<script>
  import { replace } from "svelte-spa-router";
  import { api } from "../../shared.js";
  import goto from "../lib/goto.js";
  import storeListsByMe from "../../stores/lists_by_me.js";
  import storeListsEdits from "../../stores/editor_lists_edits.js";

  import Box from "./Box.svelte";
  import ListEditTitle from "./list_edit_title.svelte";
  import ListEditDataV1 from "./list_edit_data_v1.svelte";
  import ListEditDataV2 from "./list_edit_data_v2.svelte";
  import ListEditDataV3 from "./list_edit_data_v3.svelte";
  import ListEditDataV4 from "./list_edit_data_v4.svelte";
  import ListEditDataTodo from "./list_edit_data_todo.svelte";
  import ListEditLabels from "./list_edit_labels.svelte";
  import ListEditInteractV1 from "./list_edit_interact_v1.svelte";
  import ListEditInteractV2 from "./list_edit_interact_v2.svelte";

  export let aList;

  let listTypes = {
    v1: ListEditDataV1,
    v2: ListEditDataV2,
    v3: ListEditDataV3,
    v4: ListEditDataV4
  };

  let listTypesCanInteract = {
    v1: ListEditInteractV1,
    v2: ListEditInteractV2
  };

  let renderItem = Object.keys(listTypes)
    .filter(key => aList.info.type === key)
    .reduce((notFound, key) => {
      return listTypes[key];
    }, ListEditDataTodo);

  let interactElement = Object.keys(listTypesCanInteract)
    .filter(key => aList.info.type === key)
    .reduce((notFound, key) => {
      return listTypesCanInteract[key];
    }, null);

  function cancel() {
    storeListsEdits.remove(aList.uuid);
    goto.list.view(aList.uuid);
  }

  async function remove() {
    try {
      await api.deleteList(aList.uuid);
    } catch (error) {
      alert("failed try again");
      console.error("status from server was", error);
      return;
    }

    storeListsByMe.remove(aList.uuid);
    storeListsEdits.remove(aList.uuid);
    // TODO how to remove /lists/view/:uuid as well
    replace("/list/deleted");
  }

  async function save() {
    try {
      aList = await api.updateList(aList);
    } catch (error) {
      alert("failed try again");
      console.error("status from server was", error);
      return;
    }

    try {
      storeListsEdits.remove(aList.uuid);
      storeListsByMe.update(aList);
      goto.list.view(aList.uuid);
    } catch (e) {
      alert("failed to clean up your edits");
    }
  }

  $: storeListsEdits.update(aList);
</script>

<Box>
  <ListEditTitle bind:title={aList.info.title} />
</Box>

<Box>
  <ListEditLabels bind:labels={aList.info.labels} />
</Box>

<Box>
  <svelte:component this={renderItem} bind:listData={aList.data} />
</Box>

<Box>
  <button on:click={save}>Save</button>
  <button on:click={cancel}>Cancel</button>
</Box>

<Box>
  <h1>Advanced</h1>
  <Box>
    <h2>Share</h2>
    <label>
      <input type="radio" bind:group={aList.info.shared_with} value="private" />
      Private
    </label>
    <label>
      <input type="radio" bind:group={aList.info.shared_with} value="public" />
      Public
    </label>
    <label>
      <input type="radio" bind:group={aList.info.shared_with} value="friends" />
      Friends
    </label>
  </Box>

  {#if interactElement}
    <Box>
      <h2>Interact</h2>

      <svelte:component
        this={interactElement}
        bind:interact={aList.info.interact} />
    </Box>
  {/if}

  <Box>
    <h1>Danger</h1>
    <Box>
      <button on:click={remove}>Delete this list forever</button>
    </Box>
  </Box>
</Box>
