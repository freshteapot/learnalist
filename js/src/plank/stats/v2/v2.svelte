<script>
  // TODO
  // 1) No need for store
  // 2) Handle errors
  // 3) Might need the api to know the difference between 404 and 403
  // 4) Login prompt
  // 5) Confirm login prompt redirect works
  // 6) Confirm payload based on shared + api + openapi
  // 7) Can I rework the api to not be so heavy?

  import { tick } from "svelte";

  import { PlankApi } from "../../../openapi";
  import { notify, clearNotification } from "../../../shared.js";
  import { getApi } from "../../../api.js";

  import { onMount } from "svelte";
  import Summary from "./summary.svelte";

  let loaded = false;
  let data = [];
  let debug = false;
  let hasPlankHistoryUUID = false;
  let plankUUID = "";
  let loginChallenge = false;

  onMount(async () => {
    const params = new URLSearchParams(location.search);
    const api = getApi(PlankApi);

    if (params.has("debug") && params.get("debug") === "1") {
      debug = true;
    }

    if (!params.has("plankHistory")) {
      loaded = true;
      return;
    }

    const uuid = params.get("plankHistory");
    plankUUID = uuid;
    hasPlankHistoryUUID = true;

    try {
      const response = await api.getPlankHistoryByUserRaw({ uuid });
      data = await response.value();
      loaded = true;
    } catch (e) {
      loaded = true;
      // TODO do we want 404
      if (e.status === 403) {
        loginChallenge = true;
        return;
      }
      console.log("plankHistory", uuid, e, e.status);
    }
  });

  function beforeunload(event) {
    clearNotification();
  }

  //$: history = data;
</script>

<svelte:window on:beforeunload={beforeunload} />

{#if loaded}
  {#if debug}
    <p>Has loaded {loaded}</p>
    <p>Has Plank History UUID {hasPlankHistoryUUID}</p>
    <p>Plank History UUID is {plankUUID}</p>
    <p>Do you need to login to see this list {loginChallenge}</p>
  {/if}

  {#if !loginChallenge}
    <Summary history={data} />
  {/if}
{/if}
