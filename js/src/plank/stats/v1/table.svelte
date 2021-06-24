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

    console.log(records);

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
  $: console.log(streakInfo);

  $: items = streakInfo.ranges.map((range) => {
    const start = range.start.toLocaleDateString("en-CA");
    // Seems one day off
    const end = range.end ? range.end.toLocaleDateString("en-CA") : start;
    return {
      start,
      end,
      duration: range.duration,
    };
  });
</script>

<div class="pa4">
  <div class="overflow-auto">
    <table class="f6 w-100 mw8 center" cellspacing="0">
      <thead>
        <tr>
          <th class="fw6 bb b--black-20 tl pb3 pr3 bg-white">Start</th>
          <th class="fw6 bb b--black-20 tl pb3 pr3 bg-white">End</th>
          <th class="fw6 bb b--black-20 tl pb3 pr3 bg-white">Duration</th>
        </tr>
      </thead>
      <tbody class="lh-copy">
        {#each items as range}
          <tr>
            <td class="pv3 pr3 bb b--black-20">{range.start}</td>
            <td class="pv3 pr3 bb b--black-20">{range.end}</td>
            <td class="pv3 pr3 bb b--black-20">{range.duration}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
