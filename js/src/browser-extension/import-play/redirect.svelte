<script>
  import { push } from "svelte-spa-router";

  import store from "./store.js";

  if (store.loaded()) {
    push("/overview");
  } else {
    push("/start");
  }
</script>
