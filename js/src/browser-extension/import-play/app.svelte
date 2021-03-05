<svelte:options tag={null} />

<script>
  import Router from "svelte-spa-router";
  import { replace } from "svelte-spa-router";
  import Start from "./start.svelte";
  import Overview from "./info.svelte";
  import Settings from "./settings.svelte";
  import Redirect from "./redirect.svelte";

  const routes = {
    "/overview": Overview,
    "/start": Start,
    "/settings": Settings,
    "/": Redirect,
  };
</script>

<Router {routes} on:conditionsFailed={(event) => replace("/")} />
