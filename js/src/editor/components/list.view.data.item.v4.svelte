<script>
  export let item = {
    content: "",
    url: ""
  };
</script>

<style>
  @import "../../../all.css";
</style>

<article class="dt w-100 bb b--black-05 pb2 mt2">
  <div class="dtc v-mid pl0">
    <h1 class="f6 f5-ns fw6 lh-title black mv0">{item.content}</h1>
    <a class="f6 fw4 mt0 mb0 black-40" href={item.url} target="_blank">
      {item.url}
    </a>
  </div>
</article>
