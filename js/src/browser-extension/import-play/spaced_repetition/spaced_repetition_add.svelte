<script>
    import SpacedRepetitionAdd from "./add_v2.svelte";

    let aList = JSON.parse(document.querySelector("#play-data").innerHTML);
    let listElement = document.querySelector("#list-info");
    let playElement = document.querySelector("#play");
</script>

<SpacedRepetitionAdd {aList} {listElement} {playElement} />
