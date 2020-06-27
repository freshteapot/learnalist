<script>
  import Router from "svelte-spa-router";
  import { replace } from "svelte-spa-router";
  import Nothing from "./routes/nothing.svelte";
  import TotalRecall from "./routes/total_recall_v2.svelte";
  import Slideshow from "./routes/slideshow_v2.svelte";
  import SpacedRepetition from "./routes/spaced_repetition_v2.svelte";

  const routes = {
    "/play/total_recall": TotalRecall,
    "/play/slideshow": Slideshow,
    "/interact/spaced_repetition": SpacedRepetition,
    // Catch-all, must be last
    "*": Nothing
  };
</script>

<svelte:options tag={null} />
<Router {routes} on:conditionsFailed={event => replace('/')} />
