<script>
  import { getNext, viewed } from "./api.js";

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
    console.log("Update that we saw this");
    const status = await viewed(data.uuid);
    console.log(status);
    console.log("Get next item");
    get();
  }

  async function get() {
    const response = await getNext();
    if (response.status == 200) {
      // show card
      data = response.body;
      state = "show-entry";
      return;
    }

    if (response.status == 204) {
      state = "nothing-to-see";
      // show nothing to see
      return;
    }

    // TODO
    state = "loading";
    data = null;
  }

  get();

  function showInfo(state) {
    console.log("showInfo", state);
    if (state !== "nothing-to-see") {
      listElement.style.display = "none";
      playElement.style.display = "";
      return;
    }

    listElement.style.display = "";
    playElement.style.display = "none";
  }

  $: showInfo(state);
</script>

<style>
  @import "../../../../all.css";
</style>

<svelte:options tag={null} />

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
    <p>Nothing to show</p>
  {/if}
</article>
