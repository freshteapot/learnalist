<script>
  import Modal from "./spaced_repetition_modal.svelte";

  let aList = JSON.parse(document.querySelector("#play-data").innerHTML);
  let listDataElement = document.querySelector("#list-data");
  let playElement = document.querySelector("#play");
  let data;
  let showKey = "from";
  let show = false;
  function edit(event) {
    data = event.detail.data;
    show = true;
  }

  function close() {
    data = null;
    showKey = "from";
    show = false;
  }

  function add(event) {
    console.log("Add item to spaced based learning");
    // TODO maybe make a hash out of "show", to lookup to see if unique?
    const input = {
      show: data[showKey],
      data: data,
      settings: {
        show: showKey
      },
      kind: aList.info.type
    };
    console.log(input);
    console.log("Send to server for enhanced learning");

    close();
  }
</script>

<style>
  @import "../../../../all.css";
</style>

<svelte:options tag={null} accessors={true} />

<Modal
  {aList}
  {listDataElement}
  {playElement}
  {show}
  on:add={add}
  on:edit={edit}
  on:close={close}>
  <p>
    <span>Which to show?</span>
  </p>
  <p>
    <input type="radio" bind:group={showKey} value={'from'} />
    from
  </p>
  <p>
    <input type="radio" bind:group={showKey} value={'to'} />
    to
  </p>
  <pre>{JSON.stringify(data, '', 2)}</pre>
</Modal>
