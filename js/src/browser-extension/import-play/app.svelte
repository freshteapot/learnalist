<script>
  import Router from "svelte-spa-router";
  import { replace } from "svelte-spa-router";
  import Start from "./start.svelte";
  import Overview from "./info.svelte";
  import Settings from "./settings.svelte";
  import Toolbox from "./toolbox.svelte";
  import ReadWriteRepeat from "./read-write-repeat/app.svelte";
  import Redirect from "./redirect.svelte";

  const routes = {
    "/overview": Overview,
    "/start": Start,
    "/settings": Settings,
    "/toolbox": Toolbox,
    "/toolbox/read-write-repeat": ReadWriteRepeat,
    "/": Redirect
  };
</script>

<svelte:options tag={null} />
<Router {routes} on:conditionsFailed={event => replace('/')} />
