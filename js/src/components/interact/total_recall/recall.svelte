<script>
  import { createEventDispatcher } from "svelte";

  const dispatch = createEventDispatcher();
  export let data = [];

  let state = "playing";
  let playing = false;
  // clean the inputs
  let found = [];
  let playData = [];
  let guesses = [];
  let hasChecked = false;

  playData = data.map(item => {
    return clean(item);
  });

  let leftToFind = playData.length;

  playing = true;
  let perfect = false;

  let feedback = Array(playData.length).fill("");
  let results = [];
  let attempts = 0;
  function check() {
    attempts = attempts + 1;
    hasChecked = true;
    console.log(guesses);
    results = guesses.map(input => {
      return clean(input);
    });
    // Get the unique results
    let uniques = Array.from(new Set(results));

    uniques = uniques.filter(item => playData.includes(item));

    let lookUp = uniques.map(item => {
      return {
        data: item,
        position: -1
      };
    });

    results.forEach((input, position) => {
      lookUp = lookUp.map(item => {
        // skip if already found
        if (item.position !== -1) {
          return item;
        }

        if (item.data !== input) {
          return item;
        }

        item.position = position;
        return item;
      });
    });

    // Set all to not found
    feedback = Array(playData.length).fill("notfound");

    lookUp = lookUp.map(item => {
      if (item.position === -1) {
        return item;
      }

      feedback[item.position] = "found";
      return item;
    });

    leftToFind = playData.length - uniques.length;

    if (leftToFind === 0) {
      state = "finished";
      if (attempts === 1) {
        perfect = JSON.stringify(results) === JSON.stringify(playData);
      }
      console.log("actual", JSON.stringify(playData));
      console.log("guesses", JSON.stringify(results));
    }
  }

  function playAgain() {
    dispatch("finished", {
      perfect: perfect,
      attempts: attempts,
      playAgain: true
    });
  }

  function restart() {
    dispatch("finished", {
      perfect: perfect,
      attempts: attempts,
      playAgain: false
    });
  }

  function showMe() {
    state = "show-me";
  }

  function clean(input) {
    // TODO have the UI allow for more options
    return input.toLowerCase();
  }
</script>

<style>
  @import "tachyons";
  .notfound {
    border: 4px solid #ff725c;
    border-radius: 2px;
  }

  .found {
    border: 4px solid #19a974;
    border-radius: 2px;
  }
</style>

{#if state === 'playing'}
  <p>
    How many do you remember?
    {#if hasChecked}{leftToFind} left{/if}
  </p>

  {#each playData as item, index}
    <div class="pv1">
      <input
        class="w-100 {feedback[index]}"
        disabled={feedback[index] === 'found'}
        type="text"
        placeholder=""
        bind:value={guesses[index]} />
    </div>
  {/each}
  <div class="pv1">
    <button class="br3" on:click={check}>check</button>
    <button class="br3" on:click={showMe}>I give up, show me</button>
    <button class="br3" on:click={restart}>restart</button>
  </div>
{/if}

{#if state === 'finished'}
  <p>Well done! You did it.</p>

  {#each playData as item, index}
    <div class="pv1">
      <input
        class="w-100 {feedback[index]}"
        disabled={feedback[index] === 'found'}
        type="text"
        placeholder=""
        bind:value={guesses[index]} />
    </div>
  {/each}

  {#if perfect}
    <p>Perfect recall!</p>
  {/if}
  <p>You took {attempts} attempt(s)</p>

  <div class="pv1">
    <button class="br3" on:click={playAgain}>play again</button>
    <button class="br3" on:click={restart}>restart</button>
  </div>
{/if}

{#if state === 'show-me'}
  <p>
    How many do you remember?
    {#if hasChecked}{leftToFind} left{/if}
  </p>

  {#each playData as item, index}
    <div class="pv1">
      <input
        class="w-100 found"
        disabled="true"
        type="text"
        placeholder=""
        value={item} />
    </div>
  {/each}

  <div class="pv2">
    <button class="br3" on:click={playAgain}>play again</button>
    <button class="br3" on:click={restart}>restart</button>
  </div>
{/if}
