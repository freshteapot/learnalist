<script>
  import { formatTime } from "./utils.js";
  export let timerNow = 0;
  export let settings = { showInteval: false, intervalTime: 15 };
  export let intervalTimerNow = 0;
  export let laps = 0;
</script>

<h1>{formatTime(timerNow)}</h1>
{#if settings.showIntervals}
  <p>Show interval timer ({settings.intervalTime} seconds)</p>
  <p>{formatTime(intervalTimerNow)}</p>
  <p>{laps} lap(s)</p>
{/if}

<style>
  @import "../../all.css";
</style>
