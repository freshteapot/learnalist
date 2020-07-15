<script>
  import goto from "../lib/goto.js";
  import { api } from "../../store.js";
  import { copyObject } from "../../utils/utils.js";

  import listsEdits from "../../stores/editor_lists_edits.js";
  import DataV1 from "./list_view_data_v1.svelte";
  import DataV2 from "./list_view_data_v2.svelte";
  import DataV4 from "./list_view_data_v4.svelte";
  import ItemV3 from "./list.view.data.item.v3.svelte";

  // aList object
  export let aList = {};
  export let uuid = "";
  export let info = {
    title: "",
    labels: []
  };
  export let data = [];

  let title = info.title;
  let labels = info.labels;
  let listType = info.type;
  let items = {
    v1: DataV1,
    v2: DataV2,
    v3: ItemV3,
    v4: DataV4
  };
  let renderItem = items[listType];

  function edit() {
    const edit = copyObject(aList);
    listsEdits.add(edit);
    goto.list.edit(uuid);
  }

  function view() {
    window.open(`/alist/${aList.uuid}.html`, "_blank");
  }
</script>

<style>
  @import "../../../all.css";
</style>

<div>
  <button on:click={edit}>Edit list</button>
  <button on:click={view}>View ({aList.info.shared_with})</button>
</div>

<div>
  <h1>{title}</h1>
  <p>{uuid}</p>

</div>

{#if labels.length > 0}
  <div class="nicebox">
    <ul class="list pl0">

      {#each labels as item}
        <li class="dib mr1 mb2 pl0">
          <span
            href="#"
            class="f6 f5-ns b db pa2 link dark-gray ba b--black-20">
            {item}
          </span>
        </li>
      {/each}
    </ul>
  </div>
{/if}

{#if data.length > 0}
  {#if ['v1', 'v2', 'v4'].includes(listType)}
    <svelte:component this={renderItem} {data} />
  {:else}
    <div class="nicebox">
      {#each data as item}
        <svelte:component this={renderItem} bind:item />
      {/each}
    </div>
  {/if}
{/if}
