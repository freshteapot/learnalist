<script>
  import { count, login, loggedIn, notify } from "./store.js";
  function increment() {
    count.update(n => n + 1);
  }

  function handleLogin() {
    login("/sample.html");
    notify("info", "welcome");
  }
</script>

<style>
  h1 {
    font-family: monospace;
  }
  h2 {
    color: #ff0000;
  }
</style>

<h1 on:click={increment}>Chris, count is {$count}...</h1>
<h2 on:click={handleLogin}>Login</h2>

{#if loggedIn()}
  <p>Hello</p>
{/if}
