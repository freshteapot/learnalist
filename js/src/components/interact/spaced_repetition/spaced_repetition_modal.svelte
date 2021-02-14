<svelte:options tag={null} accessors={true} />

<script>
  import { createEventDispatcher } from "svelte";

  const dispatch = createEventDispatcher();

  export let show;
  export let state;

  function handleClose() {
    dispatch("close");
  }
</script>

{#if show}
  <div class="modal-background" on:click={handleClose} />

  <div class="modal" role="dialog" aria-modal="true">
    {#if state === "edit"}
      <slot />
      <button class="br3" on:click={() => dispatch("add")}>Add</button>
    {/if}
    {#if state === "feedback"}
      <slot />
    {/if}

    <button class="br3" on:click={handleClose}>cancel</button>
  </div>
{/if}

<style>
  @import "../../../../all.css";

  .modal-background {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: rgba(0, 0, 0, 0.3);
  }

  .modal {
    position: absolute;
    left: 50%;
    top: 50%;
    width: calc(100vw - 4em);
    max-width: 32em;
    max-height: calc(100vh - 4em);
    overflow: auto;
    transform: translate(-50%, -50%);
    padding: 1em;
    border-radius: 0.2em;
    background: white;
  }
</style>
