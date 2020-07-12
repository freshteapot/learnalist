<script>
  import { link } from "svelte-spa-router";
</script>

<style>
  @import "../../../all.css";
</style>

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
</ul>
