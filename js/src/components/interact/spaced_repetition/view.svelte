<script>
  import { getNext, viewed } from "./api.js";
  import { loggedIn } from "../../../store.js";

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
      const response = await viewed(data.uuid);
      console.log("next response", response);
      get();
    } catch (error) {
      alert("Failed to talk to the server, try again");
    }
  }

  async function get() {
    if (!loggedIn()) {
      state = "nothing-to-see";
      return;
    }

    const response = await getNext();

    if (response.status == 200) {
      // show card
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

  $: get();
  $: showInfo(state);
</script>

<style>
  @import "../../../../all.css";
</style>

<svelte:options tag={null} />

{#if loggedIn()}
  <article>
    {#if state === 'show-entry'}
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
    {/if}

    {#if state === 'nothing-to-see'}
      <script>
        superstore.notifications.add(
          "info",
          "None of your entries are ready, add more?"
        );
      </script>
    {/if}

    {#if state === 'no-entries'}
      <script>
        superstore.notifications.add(
          "info",
          "Would you like to try Spaced Repetition?"
        );
      </script>
    {/if}
  </article>
{/if}
