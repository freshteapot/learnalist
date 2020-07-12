<script>
  import listsEdits from "../store/lists_edits.js";
  import Box from "../components/Box.svelte";
  import ListEdit from "../components/list_edit.svelte";

  export let params = {};

  const aList = listsEdits.find(params.uuid);
  $: exists = !!aList;
</script>

{#if exists}
  <ListEdit {aList} />
{:else}
  <p>
    list uuid is
    <span>{params.uuid}</span>
  </p>
{/if}
