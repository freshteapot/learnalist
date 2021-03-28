<script>
  import humanizeDuration from "humanize-duration";
  import { summary, streakRanges, trackRecord } from "date-streaks";
  import { monthTotals, todayTotals, totals, weekTotals } from "./helpers.js";

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
    let records = history.map((entry) => {
      // format(new Date(entry.beginningTime), "YYYY-MM-DD");
      // https://stackoverflow.com/a/38148759
      return new Date(entry.beginningTime).toLocaleDateString("en-CA");
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

  $: stats = [
    {
      name: "Today",
      value: humanizeDuration(todayTotals(history), {
        largest: 2,
        maxDecimalPoints: 2,
      }),
    },
    {
      name: "Week",
      value: humanizeDuration(weekTotals(history), {
        largest: 2,
        maxDecimalPoints: 2,
      }),
    },
    {
      name: "Month",
      value: humanizeDuration(monthTotals(history), {
        largest: 2,
        maxDecimalPoints: 2,
      }),
    },
    {
      name: "Overall",
      value: humanizeDuration(totals(history), {
        largest: 2,
        maxDecimalPoints: 2,
      }),
    },
  ];

  $: streakInfo = info(history);
  $: minutes = {
    total: history.map((e) => e.timerNow).reduce((a, b) => a + b, 0),
  };
  $: totalAttempts = history.length;
  $: streakDayRecords = streakInfo.records;
  $: streakDayRecordsDone = streakDayRecords.filter((e) => e.active).length;
</script>

<article class="pa3 pa5-ns" data-name="slab-stat-small">
  <h1 class="f6 ttu tracked">Plank Stats</h1>

  <dl class="fl fn-l w-50 dib-l w-auto-l lh-title mr5-l">
    <dd class="f6 fw4 ml0">Today</dd>
    <dd class="f3 fw6 ml0">
      {streakInfo.summary.todayInStreak ? "Done" : "Not yet"}
    </dd>
  </dl>

  <dl class="fl fn-l w-50 dib-l w-auto-l lh-title mr5-l">
    <dd class="f6 fw4 ml0">Average</dd>
    <dd class="f3 fw6 ml0">
      {humanizeDuration(minutes.total / totalAttempts, {
        largest: 2,
        maxDecimalPoints: 2,
      })}
    </dd>
  </dl>

  <h3 class="f6 ttu tracked">Time Spent</h3>
  <div class="cf">
    {#each stats as stat}
      <dl class="dib mr4">
        <dd class="f6 fw4 ml0">{stat.name}</dd>
        <dd class="f3 fw6 ml0">{stat.value}</dd>
      </dl>
    {/each}
  </div>

  <h3 class="f6 ttu tracked">Streak</h3>
  <div class="cf">
    <dl class="fl fn-l w-50 dib-l w-auto-l lh-title mr5-l">
      <dd class="f6 fw4 ml0">Current</dd>
      <dd class="f3 fw6 ml0">{streakInfo.summary.currentStreak}</dd>
    </dl>
    <dl class="fl fn-l w-50 dib-l w-auto-l lh-title mr5-l">
      <dd class="f6 fw4 ml0">Longest</dd>
      <dd class="f3 fw6 ml0">{streakInfo.summary.longestStreak}</dd>
    </dl>
  </div>

  <h3 class="f6 ttu tracked">Days</h3>
  <div class="cf">
    <dl class="fl fn-l w-50 dib-l w-auto-l lh-title mr5-l">
      <dd class="f6 fw4 ml0">Completed</dd>
      <dd class="f3 fw6 ml0">
        {streakDayRecordsDone}
      </dd>
    </dl>

    <dl class="fl fn-l w-50 dib-l w-auto-l lh-title mr5-l">
      <dd class="f6 fw4 ml0">Missed</dd>
      <dd class="f3 fw6 ml0">
        {streakDayRecords.length - streakDayRecordsDone}
      </dd>
    </dl>
  </div>

  <h3 class="f6 ttu tracked">Timeline</h3>
  <div class="cf">
    <div class="flex flex-wrap">
      {#each streakDayRecords as dayRecord}
        <div
          class="outline pa2 b--black-20 ma1"
          class:bg-green={dayRecord.active}
          class:red={!dayRecord.active}
          class:bg-red={!dayRecord.active}
        >
          &nbsp;
        </div>
      {/each}
      <div class="outline pa2 b--black-20 ma1">&lt;-Today</div>
    </div>
  </div>
</article>

<style>
  @import "../../../all.css";
</style>
