<script>
  import { formatTime } from "./utils.js";

  let details = false;
  // Promise
  export let loggedIn;
  export let entries;

  function totals(entries) {
    return entries.reduce((a, b) => a + (b["timerNow"] || 0), 0);
  }

  function formatWhen(entry) {
    return new Date(entry.beginningTime).toISOString();
  }

  function deleteEntry(entry) {
    console.log("remove", entry);
  }
  /*
  class TravellerCollection extends Array {
    sum(key) {
      return this.reduce((a, b) => a + (b[key] || 0), 0);
    }
  }
  const c = new TravellerCollection(...entries);
  const total = c.sum("timerNow");
  */

  // let entries2 = [];
  // https://github.com/sveltejs/svelte/issues/2118#issuecomment-531586875
  // $: (async () => (entries2 = await history()))();
</script>

<style>
  @import "../../all.css";
</style>

{#if loggedIn}
  <script>
    superstore.clearNotification();
  </script>
{:else}
  <script>
    superstore.notify(
      "error",
      "History is not saved, you need to login to save it"
    );
  </script>
{/if}

{#await entries}
  <p>...waiting</p>
{:then entries}
  <p>Total Planking: {formatTime(totals(entries))}</p>
  <p>Planks</p>
  <button
    class="br3"
    on:click={() => {
      details = !details;
    }}>
    Details
  </button>

  {#each entries.reverse() as entry}
    <p>
      {formatTime(entry.timerNow)}
      {#if details}({formatWhen(entry)}){/if}
      <span on:click={deleteEntry(entry)}>delete</span>
    </p>
    {#if details}
      <pre>{JSON.stringify(entry, '', 2)}</pre>
    {/if}
  {/each}

{:catch error}
  <p style="color: red">{error.message}</p>
{/await}
