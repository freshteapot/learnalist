<script>
  import Header from "./header.svelte";
  import Router from "svelte-spa-router";
  import { replace } from "svelte-spa-router";
  import Start from "./start.svelte";
  import Overview from "./info.svelte";
  import Settings from "./settings.svelte";
  import Toolbox from "./toolbox.svelte";
  import ReadWriteRepeat from "./read-write-repeat/app.svelte";
  import TotalRecall from "../../components/interact/routes/total_recall_v2.svelte";
  import Slideshow from "../../components/interact/routes/slideshow_v2.svelte";
  import SpacedRepetitionAdd from "../../components/interact/routes/spaced_repetition_v2.svelte";
  import Redirect from "./redirect.svelte";

  const routes = {
    "/overview": Overview,
    "/start": Start,
    "/settings": Settings,
    "/toolbox": Toolbox,
    "/toolbox/read-write-repeat": ReadWriteRepeat,
    "/play/total_recall": TotalRecall,
    "/play/slideshow": Slideshow,
    "/interact/spaced_repetition/add": SpacedRepetitionAdd,
    "/": Redirect
  };
</script>

<svelte:options tag={null} />
<Header />
<div id="list-info" />
<div id="play" />
<div id="play-screen" />
<Router {routes} on:conditionsFailed={event => replace('/')} />
