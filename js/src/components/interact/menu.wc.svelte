<svelte:options tag={null} />

<script>
  export let slideshow = "0";
  export let totalrecall = "0";
  export let spacedrepetition = "0";

  function showSlideshow(event) {
    window.location.hash = "#/play/slideshow";
  }

  function showTotalRecall(event) {
    window.location.hash = "#/play/total_recall";
  }

  function showSpacedRepetition(event) {
    const goto = "#/interact/spaced_repetition/add";
    if (window.location.hash !== goto) {
      window.location.hash = goto;
      return;
    }

    window.location.hash = "#/";
  }
</script>

{#if slideshow == "1"}
  <button class="br3" on:click={showSlideshow}>Slideshow</button>
{/if}

{#if totalrecall == "1"}
  <button class="br3" on:click={showTotalRecall}>Total Recall</button>
{/if}

{#if spacedrepetition == "1"}
  <button class="br3" on:click={showSpacedRepetition}>🧠 + 💪</button>
{/if}

<style>
  @import "../../../all.css";
</style>
