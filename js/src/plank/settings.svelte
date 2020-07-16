<script>
  import { createEventDispatcher } from "svelte";
  import store from "./store.js";
  import { copyObject } from "../utils/utils.js";

  const dispatch = createEventDispatcher();

  export let settings;

  function close() {
    store.settings(settings);
    dispatch("close", {
      settings
    });
  }
</script>

<style>
  @import "../../all.css";
</style>

<button class="br3" on:click={close}>Close Settings</button>

<p>
  <span>Display intervals</span>
  <input type="checkbox" bind:checked={settings.showIntervals} />
</p>
{#if settings.showIntervals}
  <p>
    <span>Interval time</span>
    <input type="number" bind:value={settings.intervalTime} max={60} min="1" />
    <span>(seconds)</span>
  </p>
{/if}
