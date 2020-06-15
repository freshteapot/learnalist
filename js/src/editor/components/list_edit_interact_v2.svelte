<script>
  import Box from "./Box.svelte";
  export let interact;
</script>

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
