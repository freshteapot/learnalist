<script>
  export let message;

  import { createEventDispatcher } from "svelte";

  const dispatch = createEventDispatcher();

  function clear() {
    dispatch("clear");
  }
</script>

<article
  class="mw10 bt bw3 b--red bg-washed-red mw-100"
  on:click="{() => clear()}"
  title="Click to dismiss."
>
  <h2 class="f4 br3 b--red mw-100 black-70 mv0 pv2 ph4">{message}</h2>
</article>
