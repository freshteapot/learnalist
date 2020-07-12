<script>
  import { copyObject } from "../../utils/utils.js";
  import { isDeviceMobile } from "../lib/helper.js";
  import { tap } from "@sveltejs/gestures";
  import { afterUpdate } from "svelte";

  const possibleCommands = {
    nothing: "",
    newItem: "When an item is added"
  };

  const isMobile = isDeviceMobile();
  const orderHelperText = !isMobile ? "drag and drop to swap" : "tap to swap";

  const newRow = {
    from: "",
    to: ""
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
  export let listData;
  let swapItems = copyObject(_swapItems);

  afterUpdate(() => {
    if (lastCmd === possibleCommands.newItem) {
      // This only works for V1 elements
      let nodes = itemsContainer.querySelectorAll(".item-container");
      nodes[nodes.length - 1].querySelector("input:first-child").focus();
      lastCmd = possibleCommands.nothing;
    }
  });

  function add() {
    listData = listData.concat(copyObject(newRow));
    lastCmd = possibleCommands.newItem;
  }

  function remove(listItem) {
    listData = listData.filter(t => t !== listItem);
    if (!listData.length) {
      listData = [copyObject(newRow)];
    }
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
  input:disabled {
    background: #ffcccc;
    color: #333;
  }

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
    {#each listData as listItem}
      <div class="item-container pv2 bb b--black-05">
        <div class="flex flex-column item-left">
          <input
            placeholder="from"
            bind:value={listItem.from}
            class="item item-left" />
          <input
            placeholder="to"
            bind:value={listItem.to}
            class="item item-left" />
        </div>
        <div class="flex flex-column">
          <button on:click={() => remove(listItem)} class="item">x</button>
        </div>
      </div>
    {/each}

    <button on:click={add}>New</button>

    <button on:click={removeAll}>Remove all</button>

    <button on:click={toggleSortable}>Change Order</button>
  {/if}

  {#if enableSortable}
    {#each listData as listItem, pos}
      <div
        draggable="true"
        class="dropzone item-container"
        data-index={pos}
        on:dragstart={dragstart}
        on:dragover={dragover}
        on:drop={drop}
        use:tap
        on:tap={tapHandler}>
        <input
          placeholder="from"
          class="item item-left"
          value={listItem.from}
          disabled />
        <input
          placeholder="to"
          class="item item-left"
          value={listItem.to}
          disabled />
      </div>
    {/each}

    <button on:click={toggleSortable}>
      Finished ordering? ({orderHelperText})
    </button>
  {/if}
</div>
