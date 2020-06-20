<script>
  import Router from "svelte-spa-router";
  import { replace } from "svelte-spa-router";
  import Nothing from "./routes/nothing.svelte";
  import TotalRecall from "./routes/total_recall_v1.svelte";
  import Slideshow from "./routes/slideshow_v1.svelte";

  const routes = {
    "/play/total_recall": TotalRecall,
    "/play/slideshow": Slideshow,
    // Catch-all, must be last
    "*": Nothing
  };
</script>

<svelte:options tag={null} />
<Router {routes} on:conditionsFailed={event => replace('/')} />
