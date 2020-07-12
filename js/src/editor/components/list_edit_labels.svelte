<script>
  import { copyObject } from "../../utils/utils.js";
  import { onMount } from "svelte";

  let labelElement;

  export let labels = [];

  function add(event) {
    if (labelElement.value === "") {
      return;
    }
    labels = labels
      .filter(f => f !== labelElement.value)
      .concat([labelElement.value]);
    labelElement.value = "";
    labelElement.focus();
  }

  function edit(input) {
    labelElement.value = input;
    labelElement.focus();
  }

  function remove() {
    labels = labels.filter(t => t !== labelElement.value);
    labelElement.value = "";
    labelElement.focus();
  }

  let enableSortable = false;
  function toggleSortable() {
    enableSortable = enableSortable ? false : true;
  }
</script>

<style>
  .draggableContainer {
    outline: none;
  }

  input:disabled {
    background: #fff;
    color: #333;
  }

  .container {
    display: flex;
  }

  span {
    text-decoration: underline;
  }

  span + span {
    margin-left: 0.5em;
  }
</style>

{#if !enableSortable}
  <div>
    <input bind:this={labelElement} placeholder="Label" />

    <button on:click={add}>Add</button>
    {#if labelElement && labelElement.value !== ''}
      <button on:click={remove}>x</button>
    {/if}

  </div>
  <div class="container">
    {#each labels as label}
      <span class="item" on:click={() => edit(label)}>{label}</span>
    {/each}
  </div>
{/if}
