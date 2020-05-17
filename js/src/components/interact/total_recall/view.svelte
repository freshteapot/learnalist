<script>
  import { createEventDispatcher } from "svelte";

  const dispatch = createEventDispatcher();

  export let data = [];
  // in milliseconds
  export let speed = 1000;
  let index = 0;
  const size = data.length - 1;
  const cancel = () => {
    clearInterval(timeout);
  };

  const timeout = setInterval(() => {
    show = data[index];

    index = index + 1;
    if (index <= size) {
      return;
    }
    cancel();
    dispatch("finished");
  }, speed);

  $: show = data[index];
</script>

<style>
  @import "tachyons";
</style>

<blockquote class="athelas ml0 mt4 pl4 black-90 bl bw2 b--black">
  <p class="f3 lh-copy">{show}</p>
</blockquote>
