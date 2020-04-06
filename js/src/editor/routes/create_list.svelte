<script>
  import goto from "../lib/goto.js";
  import cache from "../lib/cache.js";
  import { postList } from "../../api.js";
  import myLists from "../store/lists_by_me";
  import listsEdits from "../store/lists_edits.js";
  import { push } from "svelte-spa-router";
  import ErrorBox from "../components/error_box.svelte";
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
  let message;

  function clearMessage() {
    message = null;
  }

  async function handleSubmit() {
    if (title === "") {
      message = "Title cant be empty";
      return;
    }

    if (!selected) {
      message = "Pick a list type";
      return;
    }

    const response = await postList(title, selected);
    if (response.status === 201) {
      const aList = response.body;
      const uuid = aList.uuid;

      listsEdits.add(aList);
      myLists.add(aList);

      goto.list.edit(uuid);
      return;
    }
    message = response.body.message;
  }

  function init(el) {
    el.focus();
  }
</script>

<div class="pv0 mw100">
  <div class="flex items-center justify-center pa1 bg-light-red pv3">
    <svg
      class="w1"
      data-icon="info"
      viewBox="0 0 24 24"
      style="fill:currentcolor">
      <title>info icon</title>
      <path
        d="M11 15h2v2h-2v-2zm0-8h2v6h-2V7zm.99-5C6.47 2 2 6.48 2 12s4.47 10 9.99
        10C17.52 22 22 17.52 22 12S17.52 2 11.99 2zM12 20c-4.42
        0-8-3.58-8-8s3.58-8 8-8 8 3.58 8 8-3.58 8-8 8z" />
    </svg>
    <span class="lh-title ml3">
      Some info that you want to call attention to.
    </span>
  </div>
  {#if message}
    <ErrorBox {message} on:clear={clearMessage} />
  {/if}

  <section class="center pa3 ph1-ns">
    <h1 class="f4 br3 b--yellow black-70 mv0 pv2 ph4">Create a list</h1>

    <form class="pa4 black-80" on:submit|preventDefault={handleSubmit}>
      <div class="measure">
        <input
          class="input-reset ba b--black-20 pa2 mb2 db w-100"
          type="text"
          aria-describedby="title-desc"
          placeholder="Title"
          bind:value={title}
          use:init />
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
  </section>
</div>
