<script>
  import { loggedIn } from "../store.js";
  export let loginurl = "/login.html";
</script>

<style>
  @import "tachyons";
</style>

<svelte:options tag={null} />

<div class="fr mt0">
  {#if loggedIn()}
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
