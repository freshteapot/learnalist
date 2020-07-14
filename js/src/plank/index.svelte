<script>
  import { loggedIn, clearNotification } from "../store.js";
  import { formatTime } from "./utils.js";
  import {
    get as cacheGet,
    rm as cacheRm,
    save as cacheSave
  } from "../utils/storage.js";
  import Timer from "./timer.svelte";
  import store from "../stores/plank.js";
  import Settings from "./settings.svelte";
  import History from "./history.svelte";
  import LoginModal from "../components/login_modal.svelte";
  // Used for when the user is not logged in
  const StorageKeyPlankSavedItems = "plank.saved.items";

  let state = "plank_start";
  let settings = store.settings();

  let intervalTimer;
  let timerNow = 0;
  let intervalTimerNow = 0;
  let laps = 0;
  let entry = {};

  function loadCurrent() {
    if (!loggedIn()) {
      cacheRm(StorageKeyTodaysPlank);
      return;
    }

    store.today().then(() => {
      saveEntriesFromStorage();
      if (window.location.search.includes("login_redirect=true")) {
        console.log("Assumed redirect from login");
      }
      store.history();
    });
  }

  function startTime() {
    const beginning = new Date();
    const beginningTime = beginning.getTime();

    entry.showIntervals = settings.showIntervals;
    entry.intervalTime = settings.intervalTime;
    entry.beginningTime = beginningTime;

    let beginningTimeInterval;
    if (entry.showIntervals) {
      beginningTimeInterval = beginningTime;
    }

    intervalTimer = setInterval(() => {
      const current = new Date();
      const currentTime = current.getTime();
      timerNow = currentTime - beginningTime;
      entry.currentTime = currentTime;
      entry.timerNow = timerNow;

      if (entry.showIntervals) {
        if (intervalTimerNow > entry.intervalTime * 1000) {
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

    // We could move this
    const aList = $store.today;
    if (!aList) {
      console.error("Something has gone wrong, why is there no list");
      return;
    }

    const items = cacheGet(StorageKeyPlankSavedItems, []);
    if (items.length == 0) {
      return;
    }

    aList.data.push(...items);
    store
      .record(aList)
      .then(() => {
        cacheSave(StorageKeyPlankSavedItems, []);
      })
      .catch(error => {
        console.error("saveEntriesFromStorage", error);
      });

    entry = {};
    state = "plank_start";
  }

  function shouldResetForStart(state) {
    if (state !== "plank_start") {
      return;
    }
    clearNotification();
  }

  $: loadCurrent();
  $: shouldResetForStart(state);

  // TODO highlight there are items to be saved.
  console.log("Plank index");
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
    <Timer {timerNow} {settings} {intervalTimerNow} {laps} />

    <button class="br3" on:click={stop}>Stop</button>
  {/if}

  {#if state.startsWith('plank_summary')}
    <Timer {timerNow} {settings} {intervalTimerNow} {laps} />

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
      {settings}
      on:close={event => {
        settings = event.detail.settings;
        state = 'plank_start';
      }}
      on:cancel={() => {
        state = 'plank_start';
      }} />
  {/if}

  {#if state === 'history'}
    <History />
    <button
      class="br3"
      on:click={() => {
        state = 'plank_start';
      }}>
      Close
    </button>
  {/if}
</div>
