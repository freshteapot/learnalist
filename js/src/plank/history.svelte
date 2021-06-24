<script>
  // TODO visualise with colour the daily, weekly, monthly
  // TODO visualise days in a row doing a plank
  import dayjs from "dayjs";
  import en from "dayjs/locale/en";

  import { createEventDispatcher } from "svelte";
  import store from "./store.js";
  import { loggedIn, notify } from "../shared.js";
  import { formatTime } from "./utils.js";
  import LoginModal from "../components/login_modal.svelte";
  import {
    totals,
    todayTotals,
    weekTotals,
    monthTotals,
  } from "./stats/v1/helpers";

  dayjs.locale({
    ...en,
    weekStart: 1,
  });

  const dispatch = createEventDispatcher();

  const error = store.error;

  let loginNag = true;
  const loginNagMessageDefault =
    "History is not saved, you need to login to save it";
  let loginNagMessage = loginNagMessageDefault;

  function close() {
    dispatch("close");
  }

  function closeLoginModal() {
    loginNag = false;
  }

  function deleteEntry(entry) {
    if (!loggedIn()) {
      // TODO make this a modal?
      loginNagMessage = "You can only delete entries, if logged in";
      loginNag = true;
      return;
    }

    store.deleteRecord(entry);
  }

  function showError(error) {
    if (error !== "") {
      notify("error", error);
    }
  }

  let showLoginNag = false;

  // TODO handle when the dates are wrong or empty
  $: showError($error);

  $: showLoginNag = loginNag && !loggedIn();

  $: stats = [
    {
      name: "Today",
      value: formatTime(todayTotals($store.history)),
    },
    {
      name: "Week",
      value: formatTime(weekTotals($store.history)),
    },
    {
      name: "Month",
      value: formatTime(monthTotals($store.history)),
    },
    {
      name: "Overall",
      value: formatTime(totals($store.history)),
    },
  ];
</script>

<button class="br3" on:click={close}>Close History</button>

<article class="pa3 w-100 center" data-name="slab-stat">
  {#each stats as stat}
    <dl class="dib mr4">
      <dd class="f6 f5-ns b ml0">{stat.name}</dd>
      <dd class="f3 f2-ns b ml0">{stat.value}</dd>
    </dl>
  {/each}
</article>

{#if $store.history.length > 0}
  <div class="pa0">
    <div class="overflow-auto">
      <table class="f6 w-100 mw8 center" cellspacing="0">
        <thead>
          <tr>
            <th class="fw6 bb b--black-20 tl pb3 pr3 bg-white">Day</th>
            <th class="fw6 bb b--black-20 tl pb3 pr3 bg-white">Duration</th>
            <th class="fw6 bb b--black-20 tl pb3 pr3 bg-white">When</th>
            <th class="fw6 bb b--black-20 tl pb3 pr3 bg-white">Action</th>
          </tr>
        </thead>
        <tbody class="lh-copy">
          {#each $store.history as entry}
            <tr>
              <td class="pv3 pr3 bb b--black-20">
                {dayjs(entry.beginningTime).format("YY-MM-DD")}
              </td>
              <td class="pv3 pr3 bb b--black-20">
                {formatTime(entry.timerNow)}
              </td>
              <td class="pv3 pr3 bb b--black-20">
                {new Date(entry.beginningTime).toLocaleTimeString()}
              </td>
              <td class="pv3 pr3 bb b--black-20">
                <span on:click={() => deleteEntry(entry)}>delete</span>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  </div>
{/if}

{#if showLoginNag}
  <LoginModal on:close={closeLoginModal}>
    <p>{loginNagMessage}</p>
  </LoginModal>
{/if}

<style>
  @import "../../all.css";
</style>
