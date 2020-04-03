<script>
  import cache from "./lib/cache.js";
  import Router from "svelte-spa-router";
  import TopMenu from "./components/menu_top.svelte";
  import Footer from "./components/footer.svelte";
  import Menu from "./components/menu.svelte";
  import Box from "./components/Box.svelte";

  // Import the "link" action and the methods to control history programmatically from the same module, as well as the location store
  import {
    link,
    push,
    pop,
    replace,
    location,
    querystring
  } from "svelte-spa-router";
  // Import the "active" action
  // Normally, this would be import: `import active from 'svelte-spa-router/active'`
  import active from "svelte-spa-router/active";

  // Import the list of routes
  import routes from "./routes.js";

  // Contains logging information used by tests
  //let logbox = "";

  // Handles the "conditionsFailed" event dispatched by the router when a component can't be loaded because one of its pre-condition failed
  function conditionsFailed(event) {
    // eslint-disable-next-line no-console
    // console.error('Caught event conditionsFailed', event.detail)
    // logbox += 'conditionsFailed - ' + JSON.stringify(event.detail) + '\n'

    // Replace the route
    replace("/login");
  }

  // Handles the "routeLoaded" event dispatched by the router after a route has been successfully loaded
  function routeLoaded(event) {
    // eslint-disable-next-line no-console
    // console.info('Caught event routeLoaded', event.detail)
    // logbox += 'routeLoaded - ' + JSON.stringify(event.detail) + '\n'
    cache.save(cache.keys["last.screen"], "#" + event.detail.location);
  }
</script>

<style global>
  @import "tachyons";
</style>

<TopMenu />
<main class="w-100 bt b--black-10 bg-white">
  <!-- <Menu /> -->
  <Router
    {routes}
    on:conditionsFailed={conditionsFailed}
    on:routeLoaded={routeLoaded} />
</main>

<Footer />
<main class="w-100 bt b--black-10 bg-white">

  <!-- Used for testing
<p id="logbox">{logbox}</p>
 -->
</main>
