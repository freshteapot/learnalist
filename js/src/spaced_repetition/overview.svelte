<script>
  import goto from "./goto.js";
  import { loggedIn, notify, clearNotification } from "../shared.js";
  import { getEntries } from "./api.js";
  import { clearConfiguration } from "../configuration.js";
  import LoginModal from "../components/login_modal.svelte";

  let state = "loading";
  let entries = [];

  let listElement = document.querySelector("#list-info");
  let playElement = document.querySelector("#play");

  async function get() {
    try {
      entries = await getEntries();
      state = "entries";
    } catch (error) {
      if (error.status) {
        if (error.status == 403) {
          clearConfiguration();
          state = "not-logged-in";
          return;
        }

        if (error.status == 404) {
          state = "no-entries";
          return;
        }
      }

      console.log("error", error);
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

    listElement.style.display = "none";
    playElement.style.display = "";
  }

  let loginNag = true;
  const loginNagMessageDefault =
    "You need to be logged in so we can personalise your learning experience.";
  let loginNagMessage = loginNagMessageDefault;
  let showLoginNag = false;

  function closeLoginModal() {
    goto.intro();
  }

  function checkShowLoginNag() {
    return loginNag && !loggedIn();
  }

  $: showLoginNag = loginNag && !loggedIn();

  $: get();
  $: showInfo(state);
</script>

<style>
  @import "../../all.css";
</style>

<svelte:options tag={null} />

<h1>Learn with Spaced Repetition</h1>

<p>
  <button
    class="br3"
    on:click={() => {
      clearNotification();
      goto.add();
    }}>
    Add more?
  </button>
  <button class="br3" on:click={() => goto.intro()}>Learn more</button>
</p>
<ul>
  {#each entries as entry}
    <li>
      <span>{entry.show}</span>
    </li>
  {/each}
</ul>

{#if showLoginNag}
  <LoginModal on:close={closeLoginModal}>
    <p>{loginNagMessage}</p>
  </LoginModal>
{/if}
