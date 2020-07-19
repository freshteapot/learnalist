<script>
  import { copyObject } from "../../utils/utils.js";
  import { afterUpdate } from "svelte";

  import Split from "./list_edit_data_v3_split.svelte";

  export let index;
  export let record;
  export let disabled = undefined;

  import { createEventDispatcher } from "svelte";

  const dispatch = createEventDispatcher();

  const newRow = { time: "", distance: 0, p500: "", spm: 0 };

  function addSplit() {
    record.splits.push(copyObject(newRow));
    record = record;
  }

  function removeSplit(event) {
    const splitIndex = event.detail.splitIndex;
    record.splits.splice(splitIndex, 1);

    record = record;
  }

  function remove() {
    dispatch("removeRecord", index);
  }

  function disableMe() {
    return undefined;
  }
</script>

<style>
  @import "../../../all.css";
  input:disabled {
    background: #ffcccc;
    color: #333;
  }

  .item-container {
    display: flex;
  }

  .nodrag {
    -webkit-touch-callout: none;
    -webkit-user-select: none;
    -khtml-user-select: none;
    -moz-user-select: none;
    -ms-user-select: none;
    user-select: none;
  }
</style>

<div class="item-container pv2">
  <div class="flex fl w-100">
    <div class="pa0 w-100 mr2">
      <input placeholder="when" bind:value={record.when} {disabled} />
    </div>

    <div class="pa0">
      {#if disabled === undefined}
        <button on:click={remove} class="item">x</button>
      {:else}
        <span>&nbsp;</span>
      {/if}
    </div>
  </div>
</div>

<div class="item-container pv2">
  <div class="flex flex-column fl w-100">
    <div class="flex pv0">
      <div class="w-25 pa1 mr2">
        <span>time</span>
      </div>
      <div class="w-25 pa1 mr2">
        <span>meters</span>
      </div>
      <div class="w-25 pa1 mr2">
        <span>/500m</span>
      </div>

      <div class="w-25 pa1 mr2">
        <span>s/m</span>
      </div>
      <div class="pa0">
        <span class="item pa1">&nbsp;</span>
      </div>
    </div>
  </div>
</div>
<!-- OVERALL:start -->
<div class="item-container pv1 nodrag">
  <div class="flex flex-column pv2 fl w-100 bw1 bb bt b--moon-gray">
    <Split {disabled} bind:split={record.overall} />
  </div>
</div>
<!-- OVERALL:finish -->

{#each record.splits as split, splitIndex}
  <div class="item-container pv1">
    <div class="flex flex-column fl w-100">
      <Split {disabled} {splitIndex} bind:split on:click={removeSplit} />
    </div>
  </div>
{/each}

{#if disabled === undefined}
  <div class="flex pv1">
    <button class="mr1 ph1" on:click={() => addSplit()}>Add Split</button>
  </div>
{/if}
