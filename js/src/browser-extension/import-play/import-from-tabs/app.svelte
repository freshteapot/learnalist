<script>
  /*
- Could keep track in localStorage
- Need to either add to to query string
*/
  import { text_area_resize } from "./auto_resize_height.js";
  import { api } from "../../../shared.js";

  let input = "";
  let state = "start";
  let entries = [];
  // TODO store this in local storage
  // Filter the list of this
  let ignoring = [{ from: "Norsk", to: "Engelsk" }];

  // ----
  async function nextStep() {
    const stored = await api.getSpacedRepetitionEntries();
    state = "play";
    entries = input
      .trim()
      .split("\n")
      .map(e => {
        const data = e.split("\t");
        if (!data[0]) {
          data[0] = "";
        }

        if (!data[1]) {
          data[1] = "";
        }

        return {
          from: data[0].trim(),
          to: data[1].trim()
        };
      })
      .filter(entry => {
        // Filter out what we already have
        let found = stored.some(
          lookup => JSON.stringify(lookup.data) == JSON.stringify(entry)
        );
        // Only show what we dont have
        return !found;
      })
      .filter(entry => {
        // Filter out the ones we have already flagged for ignoring
        let found = ignoring.some(
          lookup => JSON.stringify(lookup) == JSON.stringify(entry)
        );
        // Only show what we dont have
        return !found;
      });
  }

  function reset() {
    input = "";
    state = "start";
    entries = [];
  }

  function removeEntry(index) {
    entries.splice(index, 1);
    entries = entries;
  }

  // {"from": "", "to": ""}
  async function addSpacedRepetitionEntry(index, data) {
    let showKey = "from";
    const input = {
      show: data[showKey],
      data: data,
      settings: {
        show: showKey
      },
      kind: "v2"
    };

    const response = await api.addSpacedRepetitionEntry(input);
    switch (response.status) {
      case 201:
        console.log("added");
        entries.splice(index, 1);
        break;
      case 200:
        console.log("already in the system");
        entries.splice(index, 1);
        break;
      default:
        console.log("failed to add for spaced learning");
        console.log(response);
        break;
    }
  }

  // ---
</script>

<style>
  @import "./all.css";
</style>

<div class="flex flex-column w-100">
  {#if state === 'start'}
    <div class="outline pa3 mr2">
      <h1>Enter data</h1>
      <button class="br3" on:click={nextStep}>Next</button>
      <textarea
        class="db border-box hover-black w-100 ba b--black-20 pa2 br2 mb2"
        placeholder="write something here, and then practice typing it"
        bind:value={input}
        use:text_area_resize />
      <button class="br3" on:click={nextStep}>Next</button>
    </div>
  {/if}

  {#if state === 'play'}
    {#if entries.length > 0}
      {#each entries as entry, index}
        <div class="outline pa3 mr2">
          <textarea
            class="db border-box hover-black w-100 ba b--black-20 pa2 br2 mb2"
            placeholder="write something here, and then practice typing it"
            bind:value={entry.from}
            use:text_area_resize />

          <textarea
            class="db border-box hover-black w-100 ba b--black-20 pa2 br2 mb2"
            placeholder="write something here, and then practice typing it"
            bind:value={entry.to}
            use:text_area_resize />
          <button
            class="br3"
            on:click={() => addSpacedRepetitionEntry(index, entry)}>
            Add to 🧠 + 💪
          </button>
          <button class="br3" on:click={() => removeEntry(index)}>
            Remove
          </button>
        </div>
      {/each}
    {/if}
  {/if}

</div>
