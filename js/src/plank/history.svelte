<script>
  import store from "../stores/plank.js";
  import { loggedIn, notify } from "../store.js";
  import { formatTime } from "./utils.js";

  let details = false;

  function totals(entries) {
    return entries.reduce((a, b) => a + (b["timerNow"] || 0), 0);
  }

  function formatWhen(entry) {
    return new Date(entry.beginningTime).toISOString();
  }

  function deleteEntry(entry) {
    console.log("TODO remove", entry);
  }

  if (!loggedIn()) {
    notify("error", "History is not saved, you need to login to save it");
  }
</script>

<style>
  @import "../../all.css";
</style>

<p>Total Planking: {formatTime(totals($store.history))}</p>
<p>Planks</p>
<button
  class="br3"
  on:click={() => {
    details = !details;
  }}>
  Details
</button>

{#each $store.history as entry}
  <p>
    {formatTime(entry.timerNow)}
    {#if details}({formatWhen(entry)}){/if}
    <span on:click={() => deleteEntry(entry)}>delete</span>
  </p>
  {#if details}
    <pre>{JSON.stringify(entry, '', 2)}</pre>
  {/if}
{/each}
