<svelte:options tag={null} accessors={true} />

<script>
  import { createEventDispatcher } from "svelte";

  const dispatch = createEventDispatcher();
  const close = () => dispatch("close");

  function handleClose() {
    dispatch("close");
  }

  function handleLogin() {
    const searchParams = new URLSearchParams();
    const redirectUrl = window.location.href.replace(
      window.location.origin,
      ""
    );

    searchParams.set("redirect", redirectUrl);
    window.location = `/login.html?${searchParams.toString()}`;
  }
</script>

<div class="modal-background" on:click={handleClose} />
<div class="modal" role="dialog" aria-modal="true">
  <slot />
  <button class="br3" on:click={handleLogin}>Login</button>
  <button class="br3" on:click={handleClose}>cancel</button>
</div>

<style>
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
