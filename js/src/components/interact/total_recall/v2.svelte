<script>
  import Recall from "./recall.svelte";
  import View from "./view.svelte";
  import { push } from "svelte-spa-router";

  // {DomElement}
  export let listElement;
  // {DomElement}
  export let playElement;

  export let data = [];
  export let gameSize = 3;
  export let speed = 1;
  let showKey = "from";

  playElement.style.display = "";
  listElement.style.display = "none";

  function handleClose(event) {
    playElement.style.display = "none";
    listElement.style.display = "";
    push("/");
  }

  let playData = [];
  // This needs to pick the data
  let state = "not-playing";

  const shuffle = (arr, key) =>
    arr
      .map(a => [Math.random(), a])
      .sort((a, b) => a[0] - b[0])
      .map(a => a[1][key]);

  function play() {
    // reduce to 7
    // shuffle
    let temp = shuffle(data, showKey);
    playData = temp.slice(0, gameSize);
    state = "playing";
  }

  function finished(event) {
    if (event.detail.playAgain) {
      play();
      return;
    }

    state = "not-playing";
  }

  function handleFinished() {
    state = "recall";
  }

  $: maxSize = data.length;
</script>

<style>
  @import "tachyons";
</style>

<svelte:options tag={null} accessors={true} />

<article>
  <header>
    <h1 class="f2 measure">Total Recall</h1>
    <button class="br3" on:click={handleClose}>Close</button>
  </header>

  <div class="pv2">
    {#if state === 'not-playing'}
      <h1>Rules</h1>
      <p>Can you remember all the words?</p>
      <p>Can you remember the order to make it perfect?</p>
      <p>How many times do you need to check?</p>

      <p>
        <span>How many to recall?</span>
        <input type="number" bind:value={gameSize} max={maxSize} min="1" />
      </p>

      <p>
        <span>How fast? (seconds)</span>
        <input type="number" bind:value={speed} max={5} min="1" />
      </p>

      <p>
        <span>Which to show?</span>
      </p>
      <p>
        <input type="radio" bind:group={showKey} value={'from'} />
        from
      </p>
      <p>
        <input type="radio" bind:group={showKey} value={'to'} />
        to
      </p>
      <pre>{JSON.stringify(data.slice(0, 2), '', 2)}</pre>
      <button class="br3" on:click={play}>Are you ready to play?</button>
    {/if}

    {#if state === 'playing'}
      <View data={playData} on:finished={handleFinished} speed={speed * 1000} />
    {/if}

    {#if state === 'recall'}
      <Recall data={playData} on:finished={finished} />
    {/if}

  </div>

</article>
