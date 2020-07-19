<script>
  import { getNext, viewed } from "./api.js";
  import { loggedIn, notify, clearNotification } from "../../../shared.js";

  let state = "loading";
  let show;
  let data;

  let listElement = document.querySelector("#list-info");
  let playElement = document.querySelector("#play");

  function flipIt() {
    if (show) {
      show = false;
      return;
    }

    show = true;
  }

  async function next() {
    try {
      clearNotification();
      const response = await viewed(data.uuid);
      await get();
    } catch (error) {
      notify(
        "error",
        "Something went wrong talking to the server, please refresh the page",
        true
      );
    }
  }

  async function get() {
    try {
      if (!loggedIn()) {
        state = "nothing-to-see";
        return;
      }

      const response = await getNext();

      if (response.status == 200) {
        // show card
        show = false;
        data = response.body;
        state = "show-entry";
        return;
      }

      if (response.status == 204) {
        state = "nothing-to-see";
        return;
      }

      if (response.status == 404) {
        state = "no-entries";
        return;
      }

      state = "loading";
      data = null;
    } catch (error) {
      notify(
        "error",
        "Something went wrong talking to the server, please refresh the page",
        true
      );
    }
  }

  function showInfo(state) {
    if (state === "loading") {
      listElement.style.display = "none";
      playElement.style.display = "none";
      return;
    }

    if (state === "show-entry") {
      listElement.style.display = "none";
      playElement.style.display = "";
      return;
    }

    listElement.style.display = "";
    playElement.style.display = "none";
  }

  function showMessage(state) {
    if (state === "nothing-to-see") {
      notify("info", "None of your entries are ready, add more?", true);
      return;
    }

    if (state === "no-entries") {
      notify("info", "Would you like to try Spaced Repetition?", true);
      return;
    }
  }

  $: get();
  $: showInfo(state);

  $: showMessage(state);
</script>

<style>
  @import "../../../../all.css";
</style>

<svelte:options tag={null} />

{#if loggedIn() && state === 'show-entry'}
  <article>
    <header>
      <button class="br3" on:click={flipIt}>Flip</button>
      <button class="br3" on:click={next}>Next</button>
    </header>
    <blockquote class="athelas ml0 mt4 pl4 black-90 bl bw2 b--black">
      {#if !show}
        <h1>{data.show}</h1>
      {/if}

      {#if show}
        <pre>{JSON.stringify(data, '', 2)}</pre>
      {/if}
    </blockquote>
  </article>
{/if}
