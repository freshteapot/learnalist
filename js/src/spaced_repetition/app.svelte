<script>
  import Router from "svelte-spa-router";
  import { replace } from "svelte-spa-router";

  import View from "../components/interact/spaced_repetition/view.svelte";
  import Add from "./add.svelte";

  const routes = {
    "/add": Add,
    "*": View
  };
</script>

<svelte:options tag={null} />
<Router {routes} on:conditionsFailed={event => replace('/')} />
