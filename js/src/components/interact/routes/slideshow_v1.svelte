<script>
  import Slideshow from "../slideshow/v1.svelte";
  let aList = JSON.parse(document.querySelector("#play-data").innerHTML);
  let listElement = document.querySelector("#list-info");
  let playElement = document.querySelector("#play");
</script>

<Slideshow {aList} {listElement} {playElement} />
