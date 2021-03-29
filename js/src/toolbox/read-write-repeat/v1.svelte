<script>
  import { text_area_resize } from "./auto_resize_height.js";

  let input = "";
  let offset = 0;
  let chars = [];
  let state = "start";

  function nextStep() {
    state = "play";
    chars = input.trim().split("");
  }

  function reset() {
    input = "";
    offset = 0;
    chars = [];
    state = "start";
  }

  function repeat() {
    offset = 0;
    state = "play";
  }

  function handleKeydown(event) {
    if (state !== "play") {
      if (state === "finished") {
        repeat();
      }
      return;
    }

    const key = event.key;
    console.log(
      "keydown",
      key,
      `${chars[offset]}`,
      chars,
      event.keyCode,
      chars[offset].charCodeAt(0)
    );
    // Handle going back
    if (key === "Backspace") {
      offset--;
      if (offset < 0) {
        offset = 0;
      }
      event.preventDefault();
      return;
    }
    // ↵
    if (event.keyCode == 13 && chars[offset] == "↵") {
      console.log("match");
    }

    if (chars[offset] === key) {
      offset++;
      if (chars[offset] === undefined) {
        state = "finished";
      }
    }
  }

  $: mark = chars.slice(0, offset);
  $: suffix = chars.slice(offset);
</script>

<svelte:window on:keydown={handleKeydown} />
<div class="flex flex-column w-100">
  {#if state === "start"}
    <div class="outline pa3 mr2">
      <h1>What shall we train today?</h1>
      <textarea
        class="db border-box hover-black w-100 ba b--black-20 pa2 br2 mb2"
        placeholder="write something here, and then practice typing it"
        bind:value={input}
        use:text_area_resize
      />
      <button class="br3" on:click={nextStep}>Next</button>
    </div>
  {/if}

  {#if state === "play"}
    <div class="outline pa3 mr2">
      <h4>
        Start typing,
        <button class="br3" on:click={reset}>Reset</button>
        at any time
      </h4>
      <p>
        {#if mark.length > 0}
          {#each mark as w}
            <mark class="fl pre bg-light-green">{w}</mark>
          {/each}
        {/if}

        {#if suffix.length > 0}
          {#each suffix as w}
            <span class="pre">{w}</span>
          {/each}
        {/if}
      </p>
    </div>
  {/if}

  {#if state === "finished"}
    <div class="outline pa3 mr2">
      <h1>Done!</h1>
      <p>{input}</p>
      <button class="br3" on:click={reset}>Reset</button>
      <button class="br3" on:click={repeat}>Repeat</button>
      (or type to start again)
    </div>
  {/if}
</div>

<style>
  @import "../../../all.css";
</style>
