<script>
  import { formatTime } from "./utils.js";
  import { AlistInputFromJSON, Configuration, DefaultApi } from "../openapi";
  export let entries;
  let details = false;

  class TravellerCollection extends Array {
    sum(key) {
      return this.reduce((a, b) => a + (b[key] || 0), 0);
    }
  }

  const c = new TravellerCollection(...entries);
  const total = c.sum("timerNow");
  const total2 = entries.reduce((a, b) => a + (b["timerNow"] || 0), 0);

  function fakeNewList() {
    var config = new Configuration({
      accessToken: "e0cfb872-37e6-47a8-9574-2bbdf1f306ca"
    });
    var api = new DefaultApi(config);

    const aList = {
      alistInput: AlistInputFromJSON({
        data: [
          "monday",
          "tuesday",
          "wednesday",
          "thursday",
          "friday",
          "saturday",
          "sunday"
        ],
        info: {
          title: "Days of the Week",
          type: "v1",
          labels: []
        }
      })
    };

    api.addList(aList).then(
      function(data) {
        console.log("API called successfully. Returned data: " + data);
        console.log(data);
        api.deleteListByUuid({ uuid: data.uuid }).then(
          function(data) {
            console.log("API called successfully. Returned data: " + data);
            console.log(data);
          },
          function(error) {
            console.error(error);
          }
        );
      },
      function(error) {
        console.error(error);
      }
    );
  }
</script>

<style>
  @import "../../all.css";
</style>

<p>Total Planking: {formatTime(total)}</p>
<p>Total Planking: {formatTime(total2)}</p>

<p>Planks</p>
{#each entries.reverse() as entry}
  <p>{formatTime(entry.timerNow)}</p>
  {#if details}
    <pre>{JSON.stringify(entry, '', 2)}</pre>
  {/if}
{/each}

<button class="br3" on:click={fakeNewList}>Stop</button>
