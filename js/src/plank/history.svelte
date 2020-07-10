<script>
  import { formatTime } from "./utils.js";
  export let entries;
  let details = false;

  class TravellerCollection extends Array {
    sum(key) {
      return this.reduce((a, b) => a + (b[key] || 0), 0);
    }
  }

  const c = new TravellerCollection(...entries);
  const total = c.sum("timerNow");
  const total2 = entries.reduce((a, b) => a + (b["timerNow"] || 0), 0);
</script>

<style>
  @import "../../all.css";
</style>

<p>Total Planking: {formatTime(total)}</p>
<p>Total Planking: {formatTime(total2)}</p>

<p>Planks</p>
{#each entries.reverse() as entry}
  <p>{formatTime(entry.timerNow)}</p>
  {#if details}
    <pre>{JSON.stringify(entry, '', 2)}</pre>
  {/if}
{/each}
