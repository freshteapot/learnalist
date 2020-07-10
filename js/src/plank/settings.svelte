<script>
  import { createEventDispatcher } from "svelte";

  const dispatch = createEventDispatcher();

  export let showIntervals;
  export let intervalTime;

  function close() {
    dispatch("close");
  }
</script>

<style>
  @import "../../all.css";
</style>

<p>
  <span>Display intervals</span>
  <input type="checkbox" bind:checked={showIntervals} />
</p>
{#if showIntervals}
  <p>
    <span>Interval time</span>
    <input type="number" bind:value={intervalTime} max={60} min="1" />
    <span>(seconds)</span>
  </p>
{/if}

<button class="br3" on:click={close}>Close</button>
