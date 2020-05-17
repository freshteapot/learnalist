<script>
  import Recall from "./recall.svelte";
  import View from "./view.svelte";
  export let data = [];

  let gameSize = 3;
  let speed = 2;
  let showKey = "from";

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
</script>

<style>
  .box {
    border: 1px solid #aaa;
    border-radius: 2px;
    padding: 1em;
    margin: 0 0 1em 0;
    top: 50%;
    left: 50%;
    position: relative;
    text-align: center;
  }
</style>

<div class="box">

  {#if state === 'not-playing'}
    <h1>Rules</h1>
    <p>Can you remember all the words?</p>
    <p>Can you remember the order to make it perfect?</p>
    <p>How many times do you need to check?</p>

    <p>
      <span>How many to recall?</span>
      <input type="number" bind:value={gameSize} max={data.length} min="1" />
    </p>

    <p>
      <span>How fast? (seconds)</span>
      <input type="number" bind:value={speed} max={5} min="1" />
    </p>

    <pre>{JSON.stringify(data.slice(0, 2), '', 2)}</pre>
    <p>
      <span>Which to show?</span>

      <label>
        <input type="radio" bind:group={showKey} value={'from'} />
        From
      </label>

      <label>
        <input type="radio" bind:group={showKey} value={'to'} />
        To
      </label>
    </p>

    <button on:click={play}>Are you ready to play?</button>
  {/if}

  {#if state === 'playing'}
    <View data={playData} on:finished={handleFinished} speed={speed * 1000} />
  {/if}

  {#if state === 'recall'}
    <Recall data={playData} on:finished={finished} />
  {/if}

</div>
