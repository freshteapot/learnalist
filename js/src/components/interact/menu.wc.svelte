<script>
  export let slideshow = "0";
  export let totalrecall = "0";

  function showSlideshow(event) {
    window.location.hash = "#/play/slideshow";
  }

  function showTotalRecall(event) {
    window.location.hash = "#/play/total_recall";
  }
</script>

<style>
  @import "../../../all.css";
</style>

<svelte:options tag={null} />
{#if slideshow == '1'}
  <button class="br3" on:click={showSlideshow}>Slideshow</button>
{/if}

{#if totalrecall == '1'}
  <button class="br3" on:click={showTotalRecall}>Total Recall</button>
{/if}
