<script>
  export let title;
</script>

<style>
  .container {
    display: flex;
  }

  input {
    display: flex;
    flex-grow: 1; /* Set the middle element to grow and stretch */
  }
</style>

<div class="container">
  <input placeholder="Title" bind:value={title} />
</div>
