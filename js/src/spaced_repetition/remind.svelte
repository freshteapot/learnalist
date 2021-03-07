<svelte:options tag={null} />

<script>
  import { getNext, viewed } from "./api.js";
  import goto from "./goto.js";
  import { notify, clearNotification } from "../shared.js";
  import { clearConfiguration } from "../configuration.js";

  let state = "loading";
  let show;
  let data;

  let wrapper;

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
      await viewed(uuid, action);
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
      const response = await getNext();

      if (response.status == 403) {
        clearConfiguration();
        goto.overview();
        return;
      }

      if ([204, 404].includes(response.status)) {
        goto.overview();
        return;
      }

      if (response.status == 200) {
        // show card
        show = false;
        data = response.body;
        state = "show-entry";
        return;
      }

      state = "loading";
      data = null;
    } catch (error) {
      console.log("error", error);
      notify(
        "error",
        "Something went wrong talking to the server, please refresh the page",
        true
      );
    }
  }

  function showEntry(state) {
    if (state === "loading") {
      listElement.style.display = "none";
      playElement.style.display = "none";
      return;
    }

    listElement.style.display = "none";
    playElement.style.display = "";
  }

  function showMessage(state) {
    if (state === "nothing-to-see") {
      notify("info", "None of your entries are ready", true);
      return;
    }

    if (state === "no-entries") {
      notify("info", "Would you like to try Spaced Repetition?", true);
      return;
    }
  }

  function handleKeydown(event) {
    switch (event.code) {
      case "Space":
        // Stop normal behaviour and remove the current focus
        event.preventDefault();
        wrapper.querySelector(":focus") &&
          wrapper.querySelector(":focus").blur();
        flipIt();
        break;
    }
  }

  $: get();
  $: showEntry(state);
</script>

<svelte:window on:keydown={handleKeydown} />

{#if state === "show-entry"}
  <div bind:this={wrapper} class="flex flex-column">
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

<style>
  @import "../../all.css";
</style>
