<script>
  import { count, loggedIn, logout } from "./store.js";
  function increment() {
    count.update(n => n + 1);
  }

  function handleLogout() {
    console.log(loggedIn());
    console.log(logout());
  }
</script>

<style>
  h1 {
    font-family: monospace;
  }
</style>

<h1 on:click={increment}>This is good enough is {$count} :D</h1>

<h1 on:click={handleLogout}>Logout</h1>
