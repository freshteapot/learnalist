<script>
  import Router from "svelte-spa-router";
  import { replace } from "svelte-spa-router";

  import Remind from "./remind.svelte";
  import Add from "./add.svelte";
  import Overview from "./overview.svelte";
  import Intro from "./intro.svelte";

  const routes = {
    "/add": Add,
    "/remind": Remind,
    "/overview": Overview,
    "*": Intro
  };
</script>

<svelte:options tag={null} />
<Router {routes} on:conditionsFailed={event => replace('/')} />
