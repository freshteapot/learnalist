<script>
  import { getNext } from "./spaced-repetition/api.js";
  import { loggedIn } from "../store.js";
  export let loginurl = "/login.html";

  let poller;

  // TODO bring this to life
  let hasSpacedRepetition = false;
  function preLogout() {
    localStorage.clear();
    console.log("It should still click");
  }

  async function checkForSpacedRepetition() {
    if (window.location.pathname.indexOf("/spaced-repetition") !== -1) {
      clearInterval(poller);
      return;
    }

    const response = await getNext();

    if (![200, 204].includes(response.status)) {
      clearInterval(poller);
      return;
    }

    if (response.status == 200) {
      clearInterval(poller);
      hasSpacedRepetition = true;
    }

    // How to handle when
    // The user has no spaced learning vs not ready
    if (response.status == 204) {
      console.log("nothing to see");
    }
  }

  // TODO how to make this run straight away then every 1 minute
  async function checkForSpacedRepetitionStraightAwayThenPeriodically() {
    await checkForSpacedRepetition();
    poller = setInterval(function() {
      checkForSpacedRepetition();
    }, 60 * 1000);
  }

  checkForSpacedRepetitionStraightAwayThenPeriodically();
</script>

<style>
  @import "../../all.css";
</style>

<svelte:options tag={null} />

<div class="fr mt0">
  {#if loggedIn()}
    {#if hasSpacedRepetition}
      <a
        title="You have something to learn."
        href="/spaced-repetition.html"
        class="f6 fw6 hover-blue link black-70 ml0 mr2-l di">
        <button class="br3">🧠 + 💪</button>
      </a>
    {/if}

    <a
      title="create, edit, share"
      href="/editor.html"
      class="f6 fw6 hover-blue link black-70 ml0 mr2-l di">
      Create
    </a>
    <a
      title="Lists created by you"
      href="/lists-by-me.html"
      class="f6 fw6 hover-blue link black-70 di">
      My Lists
    </a>
    <a
      title="Logout"
      href="/logout.html"
      on:click={preLogout}
      class="f6 fw6 hover-blue link black-70 di ml3">
      Logout
    </a>
  {:else if window.location.pathname != loginurl}
    <a
      title="Click to login"
      href={loginurl}
      class="f6 fw6 hover-red link black-70 mr2 mr3-m mr4-l dib">
      Login
    </a>
  {/if}
</div>
