<script>
  import { link, location } from "svelte-spa-router";
  import { loginHelper } from "../lib/helper.js";
  import Login from "../components/login.svelte";
</script>

{#if !$loginHelper.loggedIn}
  <Login />
{/if}

{#if $loginHelper.loggedIn}
  <p>You are logged in</p>
{/if}
