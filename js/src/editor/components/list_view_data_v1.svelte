<script>
  export let data;
</script>

<style>
  @import "../../../all.css";
</style>

<ul class="lh-copy ph0 list">
  {#each data as item}
    <li class="pv3 pr3 bb b--black-20">{item}</li>
  {/each}
</ul>
