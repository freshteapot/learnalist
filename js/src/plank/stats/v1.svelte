<script>
  // TODO do I want the streak to be calculated via events?
  import { loggedIn, notify, clearNotification } from "../../shared.js";

  import store from "../store.js";
  import Summary from "./summary.svelte";
  import Table from "./table.svelte";

  import { onMount } from "svelte";

  onMount(async () => {
    if (!loggedIn()) {
      return;
    }

    await store.history();
  });

  function checkLogin() {
    if (!loggedIn()) {
      notify("info", "You need to login to see your summary", true);
    }
  }

  function beforeunload(event) {
    clearNotification();
  }

  $: checkLogin();
  $: isLoggedIn = loggedIn();
  $: history = $store.history;
</script>

<svelte:window on:beforeunload={beforeunload} />

{#if !isLoggedIn}
  <p>Please login first</p>
{/if}

{#if isLoggedIn}
  <Summary {history} />
  <Table {history} />
{/if}
