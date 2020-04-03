<style>
  input:disabled {
    background: #ffcccc;
    color: #333;
  }
</style>

<script>
  import { createEventDispatcher } from "svelte";
  export let disabled = undefined;
  export let index;
  export let splitIndex;
  export let split = {
    time: "",
    distance: 0,
    p500: "",
    spm: 0
  };
  const dispatch = createEventDispatcher();
  // TODO duplicate
  const click = () => {
    split = {
      time: "",
      distance: 0,
      p500: "",
      spm: 0
    };
    dispatch("click", {
      index,
      splitIndex
    });
  };
</script>

<div class="flex pv0">
  <div class="w-25 pa0 mr2">
    <input
      placeholder="time"
      {disabled}
      bind:value="{split.time}"
      class="w-100"
    />
  </div>
  <div class="w-25 pa0 mr2">
    <input
      placeholder="distance"
      type="number"
      {disabled}
      bind:value="{split.distance}"
      class="w-100"
    />
  </div>
  <div class="w-25 pa0 mr2">
    <input
      placeholder="/500m"
      {disabled}
      bind:value="{split.p500}"
      class="w-100"
    />
  </div>
  <div class="w-25 pa0 mr2">
    <input
      placeholder="spm"
      type="number"
      {disabled}
      bind:value="{split.spm}"
      class="w-100"
    />
  </div>
  <div class="pa0">
    {#if disabled === undefined}
      <button on:click="{click}" class="item">x</button>
    {:else}
      <span>&nbsp;</span>
    {/if}
  </div>
</div>
