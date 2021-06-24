<script>
  import format from "date-fns/format";
  import startOfMonth from "date-fns/startOfMonth";
  import endOfMonth from "date-fns/endOfMonth";
  import differenceInDays from "date-fns/differenceInDays";
  import add from "date-fns/add";

  import { summary, streakRanges, trackRecord } from "date-streaks";

  export let history = [];

  const orderDates = (dates) => {
    dates = Object.entries(dates)
      .sort(function (a, b) {
        return b[0] - a[0];
      })
      .reverse();

    return dates.map((d) => {
      return {
        day: new Date(Date.parse(d[0])),
        active: d[1],
      };
    });
  };

  function info(history) {
    history.shift();
    let records = history.map((entry) => {
      // Needs to be in American format!!!
      return format(new Date(entry.beginningTime), "MM/dd/yyyy");
    });

    records = [...new Set(records)];
    return {
      summary: summary(records),
      ranges: streakRanges(records),
      records: orderDates(
        trackRecord({ dates: records, length: records.length })
      ),
    };
  }

  function currentStreakDates(streakInfo) {
    try {
      let offset = streakInfo.summary.currentStreak;
      if (
        streakInfo.summary.withinCurrentStreak &&
        !streakInfo.summary.todayInStreak
      ) {
        offset += 1;
      }
      const latestStreak = streakInfo.records
        .slice(-offset)
        .filter((e) => e.active);
      return latestStreak;
    } catch (e) {
      return [];
    }
  }

  function calculateStreakThisMonth(streakRecords) {
    // Get today
    // Get start of month
    const startDay = startOfMonth(new Date());
    const endDay = endOfMonth(new Date());
    const diff = differenceInDays(endDay, startDay);
    const days = [];

    for (let i = 0; i <= diff; i++) {
      const nextDay = add(startDay, { days: i });
      const found = streakRecords.find((r) => {
        return format(nextDay, "MM/dd/yyyy") === format(r.day, "MM/dd/yyyy");
      });

      let active = false;
      let day = nextDay;
      if (found) {
        day = found.day;
        active = true;
      }

      days.push({
        day,
        active,
      });
    }
    return days;
  }

  $: streakInfo = info(history);
  $: thisMonth = calculateStreakThisMonth(streakInfo.records);
</script>

<article class="pa3 pa5-ns" data-name="slab-stat-small">
  <h1 class="f6 ttu tracked">Plank Stats</h1>

  <h3 class="f6 ttu tracked">Streak</h3>
  <div class="cf">
    <dl class="fl fn-l w-50 dib-l w-auto-l lh-title mr5-l">
      <dd class="f6 fw4 ml0">Current</dd>
      <dd class="f3 fw6 ml0">{streakInfo.summary.currentStreak}</dd>
    </dl>
    <dl class="fl fn-l w-50 dib-l w-auto-l lh-title mr5-l">
      <dd class="f6 fw4 ml0">Today</dd>
      <dd class="f3 fw6 ml0">
        {@html streakInfo.summary.todayInStreak ? "&#10004;" : "&#10006;"}
      </dd>
    </dl>
  </div>

  <h3 class="f6 ttu tracked">This month</h3>
  <div class="cf">
    <div class="flex flex-wrap">
      {#each thisMonth as dayRecord}
        <div
          class="outline pa2 b--black-20 ma1"
          class:bg-green={dayRecord.active}
          class:red={!dayRecord.active}
          class:bg-red={!dayRecord.active}
          title={dayRecord.day.toLocaleDateString("en-CA")}
        >
          &nbsp;
        </div>
      {/each}
    </div>
  </div>
</article>

<style>
  @import "../../../../all.css";
</style>
