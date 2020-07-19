<script>
  import Box from "./Box.svelte";
  export let interact;
</script>

<style>
  @import "../../../all.css";
</style>

<Box>
  <h3>Slideshow</h3>
  <label>
    <input type="radio" bind:group={interact.slideshow} value={0} />
    Disable
  </label>

  <label>
    <input type="radio" bind:group={interact.slideshow} value={1} />
    Enable
  </label>
</Box>

<Box>
  <h3>Total Recall</h3>
  <label>
    <input type="radio" bind:group={interact.totalrecall} value={0} />
    Disable
  </label>

  <label>
    <input type="radio" bind:group={interact.totalrecall} value={1} />
    Enable
  </label>
</Box>
