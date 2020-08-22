<script>
  import Router from "svelte-spa-router";
  import { replace } from "svelte-spa-router";

  import Next from "./show_next.svelte";
  import Add from "./add.svelte";
  import View from "./view.svelte";
  import Intro from "./intro.svelte";

  const routes = {
    "/add": Add,
    "/next": Next,
    "/view": View,
    "*": Intro
  };
</script>

<svelte:options tag={null} />
<Router {routes} on:conditionsFailed={event => replace('/')} />
