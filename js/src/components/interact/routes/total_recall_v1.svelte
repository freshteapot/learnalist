<script>
  import TotalRecall from "../total_recall/v1.svelte";
  let aList = JSON.parse(document.querySelector("#play-data").innerHTML);
  let listElement = document.querySelector("#list-info");
  let playElement = document.querySelector("#play");
</script>

<TotalRecall data={aList.data} {listElement} {playElement} />
