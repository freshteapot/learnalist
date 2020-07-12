<script>
  import store from "../../stores/lists_by_me.js";
  import ListView from "../components/list.view.svelte";
  export let params = {};

  const aList = store.find(params.uuid);
  $: show = aList && aList.info && aList.data;
</script>

{#if show}
  <ListView {aList} {...aList} />
{:else}
  <p>Not found</p>
{/if}
