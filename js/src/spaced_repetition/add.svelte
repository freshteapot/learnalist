<script>
  import { notify } from "../shared.js";
  import { location, querystring } from "svelte-spa-router";

  let listElement = document.querySelector("#list-info");
  let playElement = document.querySelector("#play");

  playElement.style.display = "";
  listElement.style.display = "none";

  const qs = new URLSearchParams($querystring);

  function add() {
    console.log(item);
    // if to is empty, save as v1
    notify("info", "TODO Save");
  }

  let toggleTo = false;
  const item = {
    from: "",
    to: ""
  };
  item.from = qs.has("c") ? qs.get("c") : "";
</script>

<h1 class="f2 measure" title="Spaced Repetition">🧠 + 💪</h1>
<p>
  <input type="text" bind:value={item.from} />
</p>
<p>
  <input type="checkbox" bind:checked={toggleTo} />
  <span>Add meaning / definition</span>
</p>

{#if toggleTo}
  <p>
    <input type="text" bind:value={item.to} />
  </p>
{/if}

<button class="br3" on:click={add}>Add</button>
