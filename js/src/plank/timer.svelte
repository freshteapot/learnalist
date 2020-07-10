<script>
  import { formatTime } from "./utils.js";

  export let timerNow = 0;
  export let showIntervals = false;
  export let intervalTime = 0;
  export let intervalTimerNow = 0;
  export let laps = 0;
</script>

<h1>{formatTime(timerNow)}</h1>
{#if showIntervals}
  <p>Show interval timer ({intervalTime} seconds)</p>
  <p>{formatTime(intervalTimerNow)}</p>
  <p>{laps} lap(s)</p>
{/if}
