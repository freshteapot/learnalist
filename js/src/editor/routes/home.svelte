<script>
  import { link, replace } from "svelte-spa-router";
  import { loginHelper } from "../lib/helper.js";
  import cache from "../lib/cache.js";

  function reset() {
    cache.clear();
    loginHelper.logout();
    replace("/");
  }
</script>

<h1>The Editor</h1>

<ul>
  <li>
    <a class="f5 link black" href="/create" use:link>Create</a>
  </li>
  <li>Share</li>
  <li>
    <a class="f5 link black" href="/settings/server_information" use:link>
      Server information
    </a>
  </li>
  <li>
    <a class="f5 link black" href="#" on:click={() => reset()}>Clear cache</a>
  </li>
</ul>
