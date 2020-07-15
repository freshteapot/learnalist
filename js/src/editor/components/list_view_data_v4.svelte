<script>
  export let data;
</script>

<style>
  @import "../../../all.css";
</style>

{#each data as item}
  <blockquote class="ml0 mt4 pl3 black-90 bl bw2 b--black">
    <p class="measure-wide black-60">{item.content}</p>
    <cite class="f5 fs-normal">
      <a class="black-40 near-black dim" target="_blank" href={item.url}>
        {item.url}
      </a>
    </cite>
  </blockquote>
{/each}
