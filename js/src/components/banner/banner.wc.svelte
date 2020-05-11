<script>
  import { notifications } from "../../store.js";

  let infoIcon = `M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-6h2v6zm0-8h-2V7h2v2z`;

  let errorIcon = `M11 15h2v2h-2zm0-8h2v6h-2zm.99-5C6.47 2 2 6.48 2 12s4.47 10 9.99 10C17.52 22 22 17.52 22 12S17.52 2 11.99 2zM12 20c-4.42 0-8-3.58-8-8s3.58-8 8-8 8 3.58 8 8-3.58 8-8 8z`;

  function dismiss() {
    notifications.clear();
  }

  function getIcon(level) {
    if (level == "") {
      return "";
    }

    return level == "info" ? infoIcon : errorIcon;
  }

  $: level = $notifications.level;
  $: message = $notifications.message;
  $: show = level != "" ? true : false;
</script>

<style>
  @import "tachyons";
  .error {
    background-color: #ffdfdf;
  }
  .info {
    background-color: #96ccff;
  }
</style>

<svelte:options tag={null} />
{#if show}
  <div
    class="flex items-center justify-center pa3 navy"
    class:info={level === 'info'}
    class:error={level === 'error'}
    on:click={dismiss}>
    <svg
      class="w1"
      data-icon="info"
      viewBox="0 0 24 24"
      style="fill:currentcolor;width:2em;height:2em">
      <title>info icon</title>
      <path d={getIcon($notifications.level)} />
    </svg>
    <span class="lh-title ml3">{message}</span>
  </div>
{/if}
