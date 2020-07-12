<script>
  import myLists from "../store/lists_by_me";
  import ListView from "../components/list.view.svelte";
  export let params = {};

  const aList = myLists.find(params.uuid);
  $: show = aList && aList.info && aList.data;
</script>

{#if show}
  <ListView {aList} {...aList} />
{:else}
  <p>Not found</p>
{/if}
