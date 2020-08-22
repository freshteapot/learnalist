<script>
  import { location, querystring, push } from "svelte-spa-router";
  import { loggedIn, notify } from "../shared.js";
  import LoginModal from "../components/login_modal.svelte";
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

  let loginNag = true;
  const loginNagMessageDefault =
    "You need to be logged in so we can personalise your learning experience.";
  let loginNagMessage = loginNagMessageDefault;

  function closeLoginModal() {
    push("/");
  }

  function checkShowLoginNag() {
    console.log(loginNag && !loggedIn());
    return loginNag && !loggedIn();
  }
  let showLoginNag = false;

  // TODO handle when the dates are wrong or empty

  $: showLoginNag = loginNag && !loggedIn();
</script>

<h1>Learn with Spaced Repetition</h1>

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
<button class="br3" on:click={() => push('/overview')}>cancel</button>

{#if showLoginNag}
  <LoginModal on:close={closeLoginModal}>
    <p>{loginNagMessage}</p>
  </LoginModal>
{/if}
