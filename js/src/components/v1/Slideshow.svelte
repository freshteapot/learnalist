<script>
  import { tick } from "svelte";
  import { onMount } from "svelte";

  // {DomElement}
  let listElement;
  // {DomElement}
  let playElement;
  // learnalist aList object
  export let listdata;

  let aList = {};
  onMount(async () => {
    await tick();
    aList = JSON.parse(listdata);
  });

  let loops = 0;
  let index = -1;
  let firstTime =
    "Welcome, to beginning, click next, or use the right arrow key..";
  let show = firstTime;
  let nextTimeIsLoop = 0;

  /**
   * Start / prepare the slideshow for first usage.
   * @param {DomElement} _listView
   * @param {DomElement} _playView
   */
  export function start(_listElement, _playElement) {
    show = firstTime;
    loops = 0;
    index = -1;
    nextTimeIsLoop = 0;

    playElement = _playElement;
    listElement = _listElement;
    playElement.style.display = "";
    listElement.style.display = "none";
    window.addEventListener("keydown", handleKeydown);
  }

  function forward(event) {
    index += 1;

    if (!aList.data[index]) {
      index = 0;
      nextTimeIsLoop = 1;
    }

    if (nextTimeIsLoop) {
      loops += 1;
      nextTimeIsLoop = 0;
    }

    show = aList.data[index];
  }

  function backward() {
    index -= 1;
    if (index >= 0) {
      show = aList.data[index];
    } else {
      show = firstTime;
      index = -1;
    }
  }

  function handleClose(event) {
    window.removeEventListener("keydown", handleKeydown);
    playElement.style.display = "none";
    listElement.style.display = "";
  }

  function handleKeydown(event) {
    switch (event.code) {
      case "ArrowLeft":
        backward(event);
        break;
      case "Space":
      case "ArrowRight":
        console.log("right");
        forward(event);
        break;
      default:
        console.log(event);
        console.log(`pressed the ${event.key} key`);
        break;
    }
  }
</script>

<style>
  @import "tachyons";
</style>

<svelte:options tag={null} accessors={true} />
<article>
  <header>
    <h1 class="f2 measure">Slideshow</h1>
    <button class="br3" on:click={forward}>Next</button>
    <button class="br3" on:click={handleClose}>Close</button>
  </header>
  <blockquote class="athelas ml0 mt4 pl4 black-90 bl bw2 b--black">
    <p class="f5 f4-m f3-l lh-copy measure mt0">{show}</p>
    {#if loops > 0}
      <cite class="f6 ttu tracked fs-normal">
        - {loops} (Looped over the list)
      </cite>
    {/if}
  </blockquote>
</article>
