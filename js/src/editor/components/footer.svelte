<script>
  import { link } from "svelte-spa-router";
  import { loginHelper } from "../lib/helper.js";
</script>

<header class="flex w-100 bt b--black-10 bg-white">
  <div class="w-100 pa3 items-end">
    {#if $loginHelper.loggedIn}
      <div class="fr">
        <a
          title="Components"
          href="/logout"
          use:link
          class="f6 fw6 hover-red link black-70 mr2 mr3-m mr4-l dib"
        >
          Logout
        </a>
      </div>
    {/if}
  </div>
</header>
