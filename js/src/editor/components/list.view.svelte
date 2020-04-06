<script>
  import goto from "../lib/goto.js";
  import { getServer } from "../../api.js";
  import listsEdits from "../store/lists_edits.js";
  import ItemV1 from "./list.view.data.item.v1.svelte";
  import ItemV2 from "./list.view.data.item.v2.svelte";
  import ItemV3 from "./list.view.data.item.v3.svelte";
  import ItemV4 from "./list.view.data.item.v4.svelte";

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
    v1: ItemV1,
    v2: ItemV2,
    v3: ItemV3,
    v4: ItemV4
  };
  let renderItem = items[listType];

  function edit() {
    // This is really useful to know.
    // I dont want to store the copy.
    // TODO consider moving this into the store!
    const edit = JSON.parse(JSON.stringify(aList));
    listsEdits.add(edit);
    goto.list.edit(uuid);
  }

  function view() {
    const server = getServer();
    window.open(`${server}/alist/${aList.uuid}.html`, "_blank");
  }
</script>

<div class="pa3 pa5-ns">
  <div class="pl0 measure center">

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
      <div class="nicebox">
        {#each data as item}
          <svelte:component this={renderItem} bind:item />
        {/each}
      </div>
    {/if}
  </div>
</div>
