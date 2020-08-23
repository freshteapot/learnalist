<script>
  import dayjs from "dayjs";
  import isBetween from "dayjs/plugin/isBetween";
  import isToday from "dayjs/plugin/isToday";
  import duration from "dayjs/plugin/duration";
  import relativeTime from "dayjs/plugin/relativeTime";
  import en from "dayjs/locale/en";

  import { copyObject } from "../utils/utils.js";
  import goto from "./goto.js";
  import { loggedIn, notify, clearNotification } from "../shared.js";
  import { getEntries } from "./api.js";
  import { clearConfiguration } from "../configuration.js";
  import LoginModal from "../components/login_modal.svelte";

  dayjs.locale({
    ...en,
    weekStart: 1
  });
  dayjs.extend(isBetween);
  dayjs.extend(isToday);
  dayjs.extend(duration);
  dayjs.extend(relativeTime);

  // This file needs cleaning up.
  // Having a tally of what you have viewed would be beautiful
  // ---
  function totals(entries) {
    return entries.reduce((a, b) => a + (1 || 0), 0);
  }

  function todayTotals(entries) {
    return entries.reduce((a, b) => {
      if (!dayjs(b.settings.created).isToday()) {
        return a;
      }
      return a + (1 || 0);
    }, 0);
  }

  function weekTotals(entries) {
    const startOf = dayjs().startOf("week");
    const endOf = dayjs().endOf("week");

    return entries.reduce((a, b) => {
      const now = dayjs(b.settings.created);
      if (!now.isBetween(startOf, endOf)) {
        return a;
      }
      return a + (1 || 0);
    }, 0);
  }

  function monthTotals(entries) {
    const startOf = dayjs().startOf("month");
    const endOf = dayjs().endOf("month");

    return entries.reduce((a, b) => {
      const now = dayjs(b.settings.created);
      if (!now.isBetween(startOf, endOf)) {
        return a;
      }
      return a + (1 || 0);
    }, 0);
  }

  $: stats = [
    {
      name: "Today",
      value: todayTotals(entries)
    },
    {
      name: "Week",
      value: weekTotals(entries)
    },
    {
      name: "Month",
      value: monthTotals(entries)
    },
    {
      name: "Overall",
      value: totals(entries)
    }
  ];
  // ----

  let state = "loading";
  let entries = [];
  let filtered = [];
  let filterWith = "";
  let listElement = document.querySelector("#list-info");
  let playElement = document.querySelector("#play");

  // TODO move this into a store
  async function get() {
    try {
      entries = await getEntries();
      state = "entries";
    } catch (error) {
      if (error.status) {
        if (error.status == 403) {
          clearConfiguration();
          state = "not-logged-in";
          return;
        }

        if (error.status == 404) {
          state = "no-entries";
          return;
        }
      }

      console.log("error", error);
      notify(
        "error",
        "Something went wrong talking to the server, please refresh the page",
        true
      );
      state = "nothing-to-see";
      return;
    }
  }

  function filter(entries, filterWith) {
    filtered = copyObject(entries);
    if (filterWith == "") {
      return;
    }

    if (filterWith == "today") {
      filtered = filtered.filter(entry => {
        return dayjs(entry.settings.created).isToday();
      });
      return;
    }

    if (filterWith == "week") {
      const startOf = dayjs().startOf("week");
      const endOf = dayjs().endOf("week");
      filtered = filtered.filter(entry => {
        const now = dayjs(entry.settings.created);
        return now.isBetween(startOf, endOf);
      });
      return;
    }

    if (filterWith == "month") {
      const startOf = dayjs().startOf("month");
      const endOf = dayjs().endOf("month");
      filtered = filtered.filter(entry => {
        const now = dayjs(entry.settings.created);
        return now.isBetween(startOf, endOf);
      });
      return;
    }
  }

  function setFilter(input) {
    console.log(input);
    if (input === "Overall") {
      filterWith = "";
      return;
    }
    filterWith = input.toLowerCase();
  }

  function showInfo(state) {
    if (state === "loading") {
      listElement.style.display = "none";
      playElement.style.display = "none";
      return;
    }

    listElement.style.display = "none";
    playElement.style.display = "";
  }

  function whenNext(entry) {
    const now = dayjs();
    const when = dayjs(entry.settings.when_next);
    const timeLeft = entry.settings.when_next;

    const duration = dayjs.duration(when.diff(now));
    if (duration.asMilliseconds() < 0) {
      return "now";
    }
    return duration.humanize(true);
  }

  let details = false;
  let detailsUUID = "";
  function expand(entry) {
    if (detailsUUID === entry.uuid) {
      details = false;
      detailsUUID = "";
      return;
    }

    details = true;
    detailsUUID = entry.uuid;
  }

  let loginNag = true;
  const loginNagMessageDefault =
    "You need to be logged in so we can personalise your learning experience.";
  let loginNagMessage = loginNagMessageDefault;
  let showLoginNag = false;

  function closeLoginModal() {
    goto.intro();
  }

  function checkShowLoginNag() {
    return loginNag && !loggedIn();
  }

  $: showLoginNag = loginNag && !loggedIn();

  $: get();
  $: filter(entries, filterWith);
  $: showInfo(state);
</script>

<style>
  @import "../../all.css";
</style>

<svelte:options tag={null} />
<header class="tc">
  <p>
    <button
      class="br3"
      on:click={() => {
        clearNotification();
        goto.add();
      }}>
      Add more?
    </button>
    <button class="br3" on:click={() => goto.intro()}>Learn more</button>
  </p>
</header>

<article data-name="slab-stat">
  <h2 class="f5 mb3 mt4">Stats Created</h2>
  {#each stats as stat}
    <dl class="dib mr4" on:click={() => setFilter(stat.name)}>
      <dd class="f6 f5-ns b ml0">{stat.name}</dd>
      <dd class="f3 f2-ns b ml0">{stat.value}</dd>
    </dl>
  {/each}
</article>

<div class="pa0">
  <div class="overflow-auto">
    <table class="f6 w-100 mw8 center" cellspacing="0">
      <thead>
        <tr>
          <th class="fw6 bb b--black-20 tl pb3 pr3 bg-white">Show</th>
          <th class="fw6 bb b--black-20 tl pb3 pr3 bg-white">Remind</th>
        </tr>
      </thead>
      <tbody class="lh-copy">
        {#each filtered as entry}
          <tr on:click={() => expand(entry)}>
            <td class="pv3 pr3 bb b--black-20">
              <span>{entry.show}</span>
              {#if details && entry.uuid == detailsUUID}
                <pre>{JSON.stringify(entry, '', 2)}</pre>
              {/if}
            </td>
            <td class="pv3 pr3 bb b--black-20">
              <span title={dayjs(entry.settings.when).format()}>
                {whenNext(entry)}
              </span>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>

{#if showLoginNag}
  <LoginModal on:close={closeLoginModal}>
    <p>{loginNagMessage}</p>
  </LoginModal>
{/if}
