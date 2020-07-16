<script>
  import { notifications } from "./shared.js";

  function dismiss() {
    notifications.clear();
  }

  $: level = $notifications.level;
  $: message = $notifications.message;
</script>

{#if level != ''}
  <p>level: {level}</p>
  <p>message: {message}</p>
  <p on:click={dismiss}>x</p>
{/if}
