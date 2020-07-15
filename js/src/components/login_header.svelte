<script>
  import { loggedIn, api, clearConfiguration } from "../store.js";
  export let loginurl = "/login.html";

  let poller;

  let hasSpacedRepetition = false;
  let dontLookup = false;
  function preLogout() {
    clearConfiguration();
  }

  async function checkForSpacedRepetition() {
    if (!loggedIn() || dontLookup) {
      clearInterval(poller);
      return;
    }

    if (window.location.pathname.indexOf("/spaced-repetition") !== -1) {
      clearInterval(poller);
      return;
    }

    const response = await api.getSpacedRepetitionNext();

    if (![200, 204, 404].includes(response.status)) {
      clearInterval(poller);
      return;
    }

    switch (response.status) {
      case 200:
        clearInterval(poller);
        hasSpacedRepetition = true;
        break;
      case 204:
        console.log("nothing to see");
        break;
      case 404:
        dontLookup = true;
        break;
    }
  }

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
