<svelte:options tag={null} />

<script>
  import { onMount } from "svelte";
  import { location } from "svelte-spa-router";
  import { loggedIn, api } from "../shared.js";
  import { visibilityChange } from "../utils/visibilitychange.js";
  import { clearConfiguration } from "../configuration.js";
  export let loginurl = "/login.html";

  let poller;
  let hasSpacedRepetition = false;
  let dontLookup = false;

  onMount(async () => {
    document.addEventListener(visibilityChange, handleVisibilityChange, false);
  });

  function preLogout() {
    clearConfiguration();
  }

  async function checkForSpacedRepetition() {
    if (dontLookup) {
      clearInterval(poller);
      return;
    }

    const response = await api.getSpacedRepetitionNext();

    switch (response.status) {
      case 200:
        clearInterval(poller);
        hasSpacedRepetition = true;
        dontLookup = true;
        break;
      case 204:
        dontLookup = true;
        break;
      case 401:
      case 403:
        clearConfiguration();
        // Little hack to make sure the page reloads
        window.location = window.location;
        break;
      case 404:
        dontLookup = true;
        break;
      default:
        clearInterval(poller);
    }
  }

  async function checkForSpacedRepetitionStraightAwayThenPeriodically() {
    dontLookup = false;
    await checkForSpacedRepetition();
    poller = setInterval(function () {
      checkForSpacedRepetition();
    }, 60 * 1000);
  }

  function urlChangeFilterToolbox(href) {
    if (!href.includes("/toolbox/")) {
      return false;
    }

    clearInterval(poller);
    return true;
  }

  function urlChangeFilterSpacedRepetition(href, spaLocation) {
    if (href.indexOf("/spaced-repetition") === -1) {
      return false;
    }

    if (spaLocation !== "/remind") {
      return false;
    }

    clearInterval(poller);
    // Hide as we are on the page
    hasSpacedRepetition = false;
    return true;
  }

  // Based on the window href and the hash, we can watch when the page changes
  function urlChange(href, spaLocation) {
    if (!loggedIn()) {
      return;
    }

    if (urlChangeFilterToolbox(href)) {
      return;
    }

    if (urlChangeFilterSpacedRepetition(href, spaLocation)) {
      return;
    }

    checkForSpacedRepetitionStraightAwayThenPeriodically();
  }

  function handleVisibilityChange(event) {
    if (event.target.visibilityState === "visible") {
      urlChange(event.target.location.href, $location);
    }
  }

  // Could also check when we come back to the page
  $: urlChange(window.location.href, $location);
  $: showSiteMenu = window.location.href.includes("/toolbox/") === false;
</script>

<div class="fr mt0">
  {#if loggedIn()}
    {#if showSiteMenu}
      {#if hasSpacedRepetition}
        <a
          title="You have something to learn."
          href="/spaced-repetition.html#/remind"
          class="f6 fw6 hover-blue link black-70 ml0 mr2-l di"
        >
          <button class="br3">🧠 + 💪</button>
        </a>
      {/if}

      <a
        title="create, edit, share"
        href="/editor.html"
        class="f6 fw6 hover-blue link black-70 ml0 mr2-l di"
      >
        Create
      </a>
      <a
        title="Lists created by you"
        href="/lists-by-me.html"
        class="f6 fw6 hover-blue link black-70 di"
      >
        My Lists
      </a>
    {/if}
    <a
      title="Logout"
      href="/logout.html"
      on:click={preLogout}
      class="f6 fw6 hover-blue link black-70 di ml3"
    >
      Logout
    </a>
  {:else if window.location.pathname != loginurl}
    <a
      title="Click to login"
      href={loginurl}
      class="f6 fw6 hover-red link black-70 mr2 mr3-m mr4-l dib"
    >
      Login
    </a>
  {/if}
</div>

<style>
  @import "../../all.css";
</style>
