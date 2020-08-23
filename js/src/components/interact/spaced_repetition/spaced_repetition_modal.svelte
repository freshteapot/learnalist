<script>
  import { loggedIn } from "../../../shared.js";
  import { push } from "svelte-spa-router";
  import { createEventDispatcher } from "svelte";

  const dispatch = createEventDispatcher();
  const close = () => dispatch("close");

  export let show;
  export let state;

  function handleClose() {
    dispatch("close");
  }

  function handleLogin() {
    window.location = "/login.html";
  }
</script>

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

<svelte:options tag={null} accessors={true} />

{#if show}
  <div class="modal-background" on:click={handleClose} />

  <div class="modal" role="dialog" aria-modal="true">
    {#if loggedIn()}
      {#if state === 'edit'}
        <slot />
        <button class="br3" on:click={() => dispatch('add')}>Add</button>
      {/if}
      {#if state === 'feedback'}
        <slot />
      {/if}
    {:else}
      <p>You need to be logged in to use spaced repetition</p>
      <button class="br3" on:click={handleLogin}>Login</button>
    {/if}
    <button class="br3" on:click={handleClose}>cancel</button>
  </div>
{/if}
