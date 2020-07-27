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

  async function _viewed(uuid, action) {
    try {
      clearNotification();
      const response = await viewed(uuid, action);
      await get();
    } catch (error) {
      console.log("next", error);
      notify(
        "error",
        "Something went wrong talking to the server, please refresh the page",
        true
      );
    }
  }

  async function later() {
    await _viewed(data.uuid, "incr");
  }

  async function sooner() {
    await _viewed(data.uuid, "decr");
  }

  function remind() {
    notify(
      "info",
      "To be reminded sooner, click sooner. To be reminded later, click later",
      true
    );
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
      if (error.status) {
        if (error.status == 404) {
          state = "no-entries";
          return;
        }
      }

      notify(
        "error",
        "Something went wrong talking to the server, please refresh the page",
        true
      );
      state = "nothing-to-see";
      return;
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
  <div class="flex flex-column">

    <article>
      <nav>
        <button class="br3" on:click={flipIt}>Flip</button>
      </nav>
      <blockquote class="athelas ml0 mt4 pl4 black-90 bl bw2 b--black">
        {#if !show}
          <h1>{data.show}</h1>
        {/if}

        {#if show}
          <pre>{JSON.stringify(data, '', 2)}</pre>
        {/if}
      </blockquote>

      <nav class="flex justify-around">
        <div class="w-25 pa3 mr2">
          <button class="br3" on:click={sooner}>Sooner</button>
        </div>
        <div class="w-25 pa3 mr2">
          <button class="br3" on:click={remind}>Remind</button>
        </div>
        <div class="w-25 pa3 mr2">
          <button class="br3" on:click={later}>Later</button>
        </div>

      </nav>
    </article>
  </div>
{/if}
