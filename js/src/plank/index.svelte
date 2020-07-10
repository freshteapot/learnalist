<script>
  import { formatTime } from "./utils.js";
  import Timer from "./timer.svelte";
  import Settings from "./settings.svelte";
  import History from "./history.svelte";
  let state = "plank_start";
  let showIntervals = false;
  let intervalTime = 15;

  let intervalTimer;
  let timerNow = 0;
  let intervalTimerNow = 0;
  let laps = 0;
  let history = [];
  let entry = {};
  function startTime() {
    const beginning = new Date();
    const beginningTime = beginning.getTime();
    entry.showIntervals = showIntervals;
    entry.intervalTime = intervalTime;
    entry.beginningTime = beginningTime;

    let beginningTimeInterval;
    if (showIntervals) {
      beginningTimeInterval = beginningTime;
    }

    intervalTimer = setInterval(() => {
      const current = new Date();
      const currentTime = current.getTime();
      timerNow = currentTime - beginningTime;
      entry.currentTime = currentTime;
      entry.timerNow = timerNow;

      if (showIntervals) {
        if (intervalTimerNow > intervalTime * 1000) {
          const intervalBeginning = new Date();
          beginningTimeInterval = intervalBeginning.getTime();
          intervalTimerNow = 0;
          laps++;
        } else {
          intervalTimerNow = currentTime - beginningTimeInterval;
        }
      }
    }, 10);
  }

  function stopTimer() {
    clearInterval(intervalTimer);
  }

  function start() {
    state = "plank_active";
    timerNow = 0;
    laps = 0;
    intervalTimerNow = 0;
    startTime();
  }

  function stop() {
    state = "plank_summary";
    stopTimer();
  }

  function showSettings() {
    state = "settings";
  }

  function showHistory() {
    state = "history";
  }
</script>

<style>
  @import "../../all.css";
</style>

<div class="tc">
  {#if state === 'plank_start'}
    <button class="br3" on:click={showSettings}>Settings</button>
    <button class="br3" on:click={start}>Start</button>
    <button class="br3" on:click={showHistory}>History</button>
  {/if}

  {#if state === 'plank_active'}
    <Timer
      {timerNow}
      {showIntervals}
      {intervalTime}
      {intervalTimerNow}
      {laps} />

    <button class="br3" on:click={stop}>Stop</button>
  {/if}

  {#if state === 'plank_summary'}
    <Timer
      {timerNow}
      {showIntervals}
      {intervalTime}
      {intervalTimerNow}
      {laps} />

    <button
      class="br3"
      on:click={() => {
        console.log('discard');
        state = 'plank_start';
      }}>
      Discard
    </button>

    <button
      class="br3"
      on:click={() => {
        console.log('save');
        entry.intervalTimerNow = intervalTimerNow;
        entry.laps = laps;
        history.push(entry);
        entry = {};
        state = 'plank_start';
      }}>
      Save
    </button>
  {/if}

  {#if state === 'settings'}
    <Settings
      bind:showIntervals
      bind:intervalTime
      on:close={() => {
        state = 'plank_start';
      }} />
  {/if}

  {#if state === 'history'}
    <p>Show history</p>
    <History entries={history} />
    <button
      class="br3"
      on:click={() => {
        state = 'plank_start';
      }}>
      Close
    </button>
  {/if}
</div>
