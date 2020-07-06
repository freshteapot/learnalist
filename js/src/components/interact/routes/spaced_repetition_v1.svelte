<script>
  import SpacedRepetitionAdd from "../spaced_repetition/add_v1.svelte";
  let aList = JSON.parse(document.querySelector("#play-data").innerHTML);
  let listDataElement = document.querySelector("#list-data");
  let listTitleElement = document.querySelector("#list-title");
  let playElement = document.querySelector("#play");
</script>

<SpacedRepetitionAdd
  {aList}
  {listDataElement}
  {playElement}
  {listTitleElement} />
