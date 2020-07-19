<script>
  export let item = {
    when: "",
    overall: { time: "", distance: "", p500: "", spm: "" },
    splits: []
  };

  $: description = `${item.overall.distance} meters in ${item.overall.time}`;
</script>

<style>
  @import "../../../all.css";
</style>

<article class="dt w-100 bb b--black-05 pb2 mt2">
  <div class="dtc v-mid pl0">
    <h1 class="f6 f5-ns fw6 lh-title black mv0">{description}</h1>
  </div>
</article>
