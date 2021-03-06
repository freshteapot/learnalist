<svelte:options tag={null} />

<script>
    import Router from "svelte-spa-router";
    import { replace } from "svelte-spa-router";
    import SpacedRepetitionAdd from "../spaced_repetition/spaced_repetition_add.svelte";

    const routes = {
        "/spaced_repetition/add": SpacedRepetitionAdd,
    };
</script>

<Router {routes} on:conditionsFailed={(event) => replace("/")} />
