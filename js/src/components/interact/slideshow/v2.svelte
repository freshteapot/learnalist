<script>
  import { push } from "svelte-spa-router";
  import { tap } from "@sveltejs/gestures";

  // {DomElement}
  export let listElement;
  // {DomElement}
  export let playElement;
  // learnalist aList object
  // This is horrible, as it works on aList not alist
  export let aList;

  playElement.style.display = "";
  listElement.style.display = "none";

  let loops = 0;
  let index = -1;
  let firstTime = {
    from: "Welcome, to beginning,",
    to: "click next, or use the right arrow key.."
  };

  let show = firstTime;
  let nextTimeIsLoop = 0;

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
    playElement.style.display = "none";
    listElement.style.display = "";
    push("/");
  }

  function handleKeydown(event) {
    switch (event.code) {
      case "ArrowLeft":
        backward(event);
        break;
      case "Space":
      case "ArrowRight":
        forward(event);
        break;
      default:
        console.log(event);
        console.log(`pressed the ${event.key} key`);
        break;
    }
  }

  function tapHandler(event) {
    event.preventDefault();

    // Some sort of horrible when running in the chrome extension :(
    let elem = document.elementFromPoint(event.detail.x, event.detail.y);
    if (elem && elem.nodeName === "BUTTON") {
      return false;
    }

    const margin = 150;
    const width = event.target.innerWidth; // window
    const pageX = event.detail.x; // event.pageX when touchstart
    const left = 0 + pageX < margin;
    const right = width - margin < pageX;

    if (left) {
      backward(event);
      return;
    }

    if (right) {
      forward(event);
      return;
    }

    return;
  }
</script>

<style>
  @import "../../../../all.css";
</style>

<svelte:window on:keydown={handleKeydown} use:tap on:tap={tapHandler} />
<svelte:options tag={null} accessors={true} />
<article>
  <header>
    <h1>Slideshow</h1>
    <button class="br3" on:click={forward}>Next</button>
    <button class="br3" on:click={handleClose}>Close</button>
  </header>
  <blockquote class="athelas ml0 mt4 pl4 black-90 bl bw2 b--black">
    <div class="f3 lh-copy">
      <p>{show.from}</p>
      <p>{show.to}</p>
    </div>
    {#if loops > 0}
      <cite class="f6 ttu tracked fs-normal">
        - {loops} (Looped over the list)
      </cite>
    {/if}
  </blockquote>
</article>
