<script>
  export let data;
</script>

<style>
  @import "../../../all.css";
</style>

<table class="w-100" cellspacing="0">
  <thead>
    <tr>
      <th class="fw6 bb b--black-20 pb3 tl">From</th>
      <th class="fw6 bb b--black-20 pb3 tl">To</th>
    </tr>
  </thead>
  <tbody class="lh-copy">
    {#each data as item}
      <tr>
        <td class="pv3 pr3 bb b--black-20">{item.from}</td>
        <td class="pv3 pr3 bb b--black-20">{item.to}</td>
      </tr>
    {/each}
  </tbody>
</table>
