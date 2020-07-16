<script>
  import goto from "../lib/goto.js";
  import { notify, api } from "../../shared.js";
  import { focusThis } from "../../utils/utils.js";

  import storeListsByMe from "../../stores/lists_by_me";
  import storeListEdits from "../../stores/editor_lists_edits.js";
  import { push } from "svelte-spa-router";

  let title = "";
  let listTypes = [
    {
      key: "v1",
      description: "free text"
    },
    {
      key: "v2",
      description: "From -> To"
    },
    {
      key: "v4",
      description: "A url and some text"
    },
    {
      key: "v3",
      description: "Concept2 rowing machine log"
    }
  ];
  //"v1", "v2", "v4"
  let selected;

  async function handleSubmit() {
    if (title === "") {
      notify("error", "Title cant be empty");
      return;
    }

    if (!selected) {
      notify("error", "Pick a list type");
      return;
    }
    // TODO

    let aList = {
      data: [],
      info: {
        title: title,
        type: selected,
        labels: []
      }
    };

    try {
      aList = await api.addList(aList);
    } catch (error) {
      alert("failed try again");
      console.error("status from server was", error);
      notify("error", error);
      return;
    }

    storeListEdits.add(aList);
    storeListsByMe.add(aList);

    goto.list.edit(aList.uuid);
  }
</script>

<style>
  @import "../../../all.css";
</style>

<h1 class="f4 br3 b--yellow black-70 mv0 pv2 ph4">Create a list</h1>

<form class="pa4 black-80" on:submit|preventDefault={handleSubmit}>
  <div class="measure">
    <input
      class="input-reset ba b--black-20 pa2 mb2 db w-100"
      type="text"
      aria-describedby="title-desc"
      placeholder="Title"
      bind:value={title}
      use:focusThis />
  </div>

  <div class="measure">
    <fieldset class="bn">
      {#each listTypes as listType, pos}
        <div class="flex items-center mb2">
          <input
            class="mr2"
            type="radio"
            bind:group={selected}
            id="list-type-{pos}"
            name="type"
            value={listType.key} />
          <label for="list-type-{pos}" class="lh-copy">
            {listType.description}
          </label>
        </div>
      {/each}
    </fieldset>
  </div>
  <div class="measure">
    <button type="submit">Submit</button>
  </div>
</form>
