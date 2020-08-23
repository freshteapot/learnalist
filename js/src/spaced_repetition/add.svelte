<script>
  import { location, querystring } from "svelte-spa-router";
  import goto from "./goto.js";
  import { addEntry } from "./api.js";

  import { loggedIn, notify } from "../shared.js";
  import LoginModal from "../components/login_modal.svelte";
  let listElement = document.querySelector("#list-info");
  let playElement = document.querySelector("#play");

  playElement.style.display = "";
  listElement.style.display = "none";

  const qs = new URLSearchParams($querystring);

  let toggleTo = false;
  const item = {
    from: "",
    to: ""
  };
  item.from = qs.has("c") ? qs.get("c") : "";

  async function add() {
    item.from = item.from.trim();
    item.to = item.to.trim();

    const input = {
      show: item.from
    };

    if (item.from === "") {
      notify("error", "Entry can not be empty");
      return;
    }

    if (item.to === "") {
      input.data = item.from;
      input.kind = "v1";
    }

    if (item.to !== "") {
      input.data = item;
      input.kind = "v2";
      input.settings = {
        show: "from"
      };
    }

    const response = await addEntry(input);

    switch (response.status) {
      case 201:
        notify("info", "Saved");
        item.from = "";
        item.to = "";
        toggleTo = false;
        break;
      case 200:
        notify("info", "Already in the system");
        item.from = "";
        item.to = "";
        toggleTo = false;
        break;
      default:
        console.log("failed to add for spaced learning");
        console.log(response);
        break;
    }
  }

  const loginNagMessageDefault =
    "You need to be logged in so we can personalise your learning experience.";
  let loginNagMessage = loginNagMessageDefault;

  function closeLoginModal() {
    goto.intro();
  }
</script>

<article class="tc">
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
  <button class="br3" on:click={() => goto.overview()}>cancel</button>
</article>

{#if !loggedIn()}
  <LoginModal on:close={closeLoginModal}>
    <p>{loginNagMessage}</p>
  </LoginModal>
{/if}
