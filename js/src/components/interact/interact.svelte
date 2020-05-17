<script>
  import Router from "svelte-spa-router";

  import TotalRecall from "./routes/total_recall.svelte";
  import Slideshow from "./routes/slideshow.svelte";
  import Nothing from "./routes/nothing.svelte";

  // Import the "link" action and the methods to control history programmatically from the same module, as well as the location store
  import { replace } from "svelte-spa-router";

  const routes = {
    "/play/total_recall": TotalRecall,
    "/play/slideshow": Slideshow,
    // Catch-all, must be last
    "*": Nothing
  };

  // Handles the "conditionsFailed" event dispatched by the router when a component can't be loaded because one of its pre-condition failed
  function conditionsFailed(event) {
    replace("/");
  }

  // Handles the "routeLoaded" event dispatched by the router after a route has been successfully loaded
  function routeLoaded(event) {
    // eslint-disable-next-line no-console
    console.info("Caught event routeLoaded", event.detail);
  }
</script>

<svelte:options tag={null} />
<Router
  {routes}
  on:conditionsFailed={conditionsFailed}
  on:routeLoaded={routeLoaded} />
