<script>
  import { copyObject } from "../../utils/utils.js";
  import { isDeviceMobile } from "../lib/helper.js";
  import { tap } from "@sveltejs/gestures";
  import { afterUpdate } from "svelte";
  import Record from "./list_edit_data_v3_record.svelte";

  export let listData;

  const possibleCommands = {
    nothing: "",
    newItem: "When an item is added"
  };

  const isMobile = isDeviceMobile();
  const orderHelperText = !isMobile ? "drag and drop to swap" : "tap to swap";

  const newRow = {
    when: "",
    overall: { time: "", distance: 0, p500: "", spm: 0 },
    splits: []
  };

  const _swapItems = {
    from: -1,
    fromElement: null,
    to: -1,
    toElement: null
  };

  let itemsContainer;
  let lastCmd = possibleCommands.nothing;

  let enableSortable = false;

  let swapItems = copyObject(_swapItems);

  afterUpdate(() => {
    if (lastCmd === possibleCommands.newItem) {
      // This only works for V1 elements
      //let nodes = itemsContainer.querySelectorAll(".item-container");
      //nodes[nodes.length - 1].querySelector("input:first-child").focus();
      lastCmd = possibleCommands.nothing;
    }
  });

  function add() {
    listData.push(copyObject(newRow));
    listData = listData;
    lastCmd = possibleCommands.newItem;
  }

  function remove(event) {
    listData.splice(event.detail.index, 1);
    if (!listData.length) {
      listData = [copyObject(newRow)];
    }
    listData = listData;
  }

  function removeAll() {
    listData = [copyObject(newRow)];
  }

  function toggleSortable(ev) {
    if (listData.length <= 1) {
      alert("nothing to swap");
      return;
    }

    enableSortable = enableSortable ? false : true;
    if (enableSortable) {
      // Reset swapItems
      swapItems = copyObject(_swapItems);
    }
  }

  function dragstart(ev) {
    swapItems = copyObject(_swapItems);
    swapItems.from = ev.target.getAttribute("data-index");
  }

  function dragover(ev) {
    ev.preventDefault();
  }

  function drop(ev) {
    ev.preventDefault();
    swapItems.to = ev.target.getAttribute("data-index");

    // We might land on the children, look up for the draggable attribute
    if (swapItems.to == null) {
      swapItems.to = ev.target
        .closest("[draggable]")
        .getAttribute("data-index");
    }

    let a = listData[swapItems.from];
    let b = listData[swapItems.to];
    listData[swapItems.from] = b;
    listData[swapItems.to] = a;
  }

  function tapHandler(ev) {
    ev.preventDefault();

    let index = ev.target.getAttribute("data-index");

    if (index === null) {
      swapItems = copyObject(_swapItems);
      return;
    }

    if (swapItems.from === -1) {
      swapItems.fromElement = ev.target;
      swapItems.fromElement.style["border-left"] = "solid green";
      swapItems.from = index;
      return;
    }

    if (swapItems.from === index) {
      swapItems.fromElement.style.border = "";
      swapItems = copyObject(_swapItems);
      return;
    }

    swapItems.to = index;
    let a = listData[swapItems.from];
    let b = listData[swapItems.to];
    listData[swapItems.from] = b;
    listData[swapItems.to] = a;

    swapItems.fromElement.style.border = "";
    swapItems.fromElement.style["border-radius"] = "0px";
    swapItems = copyObject(_swapItems);
  }
</script>

<style>
  @import "../../../all.css";
  .container {
    display: flex;
    justify-content: space-between;
    flex-direction: column;
  }

  .item-container {
    display: flex;
  }

  .item-container .item {
  }

  .item-container .item-left {
    flex-grow: 1; /* Set the middle element to grow and stretch */
    margin-right: 0.5em;
  }
</style>

<h1>Items</h1>

<div bind:this={itemsContainer}>
  {#if !enableSortable}
    {#each listData as listItem, index}
      <Record {index} bind:record={listItem} on:removeRecord={remove} />
    {/each}

    <div class="flex pv1">
      <button class="mr1 ph1" on:click={add}>New</button>
      <button class="mh1 ph1" on:click={removeAll}>Remove all</button>
      <button class="mh1 ph1" on:click={toggleSortable}>Change Order</button>

    </div>
  {/if}

  {#if enableSortable}
    {#each listData as listItem, pos}
      <div
        draggable="true"
        class="dropzone pv2 bb b--black-05"
        data-index={pos}
        on:dragstart={dragstart}
        on:dragover={dragover}
        on:drop={drop}
        use:tap
        on:tap={tapHandler}>
        <Record disabled="true" index={pos} bind:record={listItem} />
      </div>
    {/each}

    <button on:click={toggleSortable}>
      Finished ordering? ({orderHelperText})
    </button>
  {/if}
</div>
