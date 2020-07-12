<script>
  import { loggedIn, notifications } from "../store.js";
  import { formatTime } from "./utils.js";
  import {
    get as cacheGet,
    rm as cacheRm,
    save as cacheSave
  } from "../utils/storage.js";
  import Timer from "./timer.svelte";
  import { today, history as getHistory2, save } from "./api.js";
  import Settings from "./settings.svelte";
  import History from "./history.svelte";
  import LoginModal from "../components/login_modal.svelte";
  // Used for when the user is not logged in
  const StorageKeyPlankSavedItems = "plank.saved.items";
  // History of planks, maybe this is an api call
  const StorageKeyPlankHistory = "plank.history";
  const StorageKeyTodaysPlank = "plank.today";

  cacheRm(StorageKeyTodaysPlank);
  /*
  (async () => {
    const aList = await today();
    cacheSave(StorageKeyTodaysPlank, aList);
  })();
  */

  today().then(aList => {
    // TODO check, do we handle cleaning up after logged in, logged out.
    if (!aList) {
      cacheRm(StorageKeyTodaysPlank);
    }

    cacheSave(StorageKeyTodaysPlank, aList);

    saveEntriesFromStorage();
    /*
    if (window.location.search.includes("login_redirect=true")) {
      saveEntriesFromStorage();
    }
    */
  });

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

  function closeLoginModal() {
    entry = {};
    state = "plank_start";
  }

  function handleSave() {
    console.log("save");
    entry.intervalTimerNow = intervalTimerNow;
    entry.laps = laps;
    let items = cacheGet(StorageKeyPlankSavedItems, []);
    items.push(entry);
    cacheSave(StorageKeyPlankSavedItems, items);

    if (!loggedIn()) {
      state = "plank_summary_login";
      return;
    }

    saveEntriesFromStorage();
  }

  function saveEntriesFromStorage() {
    // How to notify that items are saved
    // Could always call this function and include loggedIn
    // Could always just auto save
    if (!loggedIn()) {
      return;
    }
    const aList = cacheGet(StorageKeyTodaysPlank, null);
    if (!aList) {
      console.error("Something has gone wrong, why is there no list");
      return;
    }

    const items = cacheGet(StorageKeyPlankSavedItems, []);
    if (items.length == 0) {
      return;
    }

    aList.data.push(...items);
    save(aList)
      .then(saved => {
        cacheSave(StorageKeyPlankSavedItems, []);
        cacheSave(StorageKeyTodaysPlank, saved);
      })
      .catch(error => {
        console.error("saveEntriesFromStorage", error);
      });

    entry = {};
    state = "plank_start";
  }

  async function getHistory() {
    if (loggedIn()) {
      return getHistory2();
    }
    return Promise.resolve(cacheGet(StorageKeyPlankSavedItems, []));

    const aList = cacheGet(StorageKeyTodaysPlank, null);
    if (aList) {
      return aList.data;
    }

    // If user is not logged in, show them the unsaved history
    // TODO make it clear this has not been saved, on the history screen
    if (!loggedIn()) {
      return cacheGet(StorageKeyPlankSavedItems, []);
    }
    return [];
  }

  // TODO highlight there are items to be saved.
</script>

<style>
  @import "../../all.css";
</style>

<div class="tc">
  {#if state === 'plank_start'}
    <script>
      superstore.clearNotification();
    </script>
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

  {#if state.startsWith('plank_summary')}
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

    <button class="br3" on:click={handleSave}>Save</button>

    {#if state === 'plank_summary_login'}
      <LoginModal on:close={closeLoginModal}>
        <p>Abc</p>
      </LoginModal>
    {/if}
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
    <History entries={getHistory()} loggedIn={loggedIn()} />
    <button
      class="br3"
      on:click={() => {
        state = 'plank_start';
      }}>
      Close
    </button>
  {/if}
</div>
