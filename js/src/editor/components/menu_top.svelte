<script>
  import { link, location } from "svelte-spa-router";
  import { loginHelper } from "../lib/helper.js";
</script>

<header class="flex">
  <div class="w-25 pa3 mr2" />

  <div class="w-75 pa3 items-end">
    <div class="fr mt0">
      {#if $loginHelper.loggedIn}
        <a
          title="Documentation"
          href="/create"
          use:link
          class="f6 fw6 hover-blue link black-70 mr2 mr3-m mr2-l di">
          Create
        </a>
        <a
          title="Components"
          href="/lists/by/me"
          use:link
          class="f6 fw6 hover-blue link black-70 mr2 mr3-m mr5-l di">
          Find
        </a>
      {:else if $location != '/login'}
        <a
          title="Components"
          href="/login"
          use:link
          class="f6 fw6 hover-red link black-70 mr2 mr3-m mr4-l dib">
          Login
        </a>
      {/if}
    </div>
  </div>
</header>
