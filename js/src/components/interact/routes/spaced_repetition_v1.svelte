<script>
  import { push } from "svelte-spa-router";
  import { onMount, onDestroy } from "svelte";

  export let params;
  let aList = JSON.parse(document.querySelector("#play-data").innerHTML);
  let listElement = document.querySelector("#list-info");
  let listDataElement = document.querySelector("#list-data");
  let playElement = document.querySelector("#play");
  let data;

  //listElement.style.display = "none";
  playElement.style.display = "";

  onMount(() => {
    listDataElement.addEventListener("click", handler);
  });

  onDestroy(() => {
    listDataElement.removeEventListener("click", handler);
    playElement.style.display = "none";
  });

  function handler(event) {
    const index = event.target.getAttribute("data-index");
    data = aList.data[index];
  }

  function add(event) {
    console.log("Add item to spaced based learning");
    // TODO maybe make a hash out of "show", to lookup to see if unique?
    const input = {
      show: data,
      data: data,
      kind: aList.info.type
    };
    console.log(input);
    console.log("Send to server for enhanced learning");
  }

  function handleClose(event) {
    data = null;
  }
</script>

<style>
  @import "../../../../all.css";

  .modal-background {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: rgba(0, 0, 0, 0.3);
  }

  .modal {
    position: absolute;
    left: 50%;
    top: 50%;
    width: calc(100vw - 4em);
    max-width: 32em;
    max-height: calc(100vh - 4em);
    overflow: auto;
    transform: translate(-50%, -50%);
    padding: 1em;
    border-radius: 0.2em;
    background: white;
  }

  button {
    display: block;
  }
</style>

<svelte:options tag={null} accessors={true} />
{#if data}
  <div class="modal-background" on:click={handleClose} />

  <div class="modal" role="dialog" aria-modal="true">
    <pre>{JSON.stringify(data, '', 2)}</pre>

    <button class="br3" on:click={add}>Add</button>
    <button class="br3" on:click={handleClose}>close modal</button>
  </div>
{/if}
