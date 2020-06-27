<script>
  console.log("TODO: bring it to life");
  let show;

  let data = {
    show: "Hello I am text",
    data: {
      from: "Hello",
      to: "Hello I am text"
    },
    kind: "v2"
  };

  function flipIt() {
    if (show) {
      show = null;
      return;
    }

    show = true;
  }
</script>

<style>
  @import "../../../all.css";
</style>

<svelte:options tag={null} />

<article>
  <header>
    <button class="br3" on:click={flipIt}>Flip</button>
  </header>
  <blockquote class="athelas ml0 mt4 pl4 black-90 bl bw2 b--black">
    {#if !show}
      <h1>{data.show}</h1>
    {/if}

    {#if show}
      <pre>{JSON.stringify(data, '', 2)}</pre>
    {/if}
  </blockquote>
</article>
