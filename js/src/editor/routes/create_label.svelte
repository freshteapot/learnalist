<script>
  // TODO this does not work
  import { hasWhiteSpace } from "../lib/helper.js";
  import { focusThis } from "../../utils/utils.js";
  import { notify } from "../../store.js";

  let newLabel = "";

  async function add() {
    if (newLabel === "" || hasWhiteSpace(newLabel)) {
      notify("error", "The label cannot be empty.");
      newLabel = "";
      return;
    }
    let response = await label.save(newLabel);
    if (response.status === 201 || response.status === 200) {
      notify("info", "The label cannot be empty.");
      router.showScreenLabelView(newLabel);
    } else {
      notify("error", response.body.message);
    }
  }

  async function handleSubmit() {
    if (newLabel === "" || hasWhiteSpace(newLabel)) {
      notify("error", "The label cannot be empty.");
      newLabel = "";
      return;
    }
    notify("info", "TODO add label to the system");
  }
</script>

<h1 class="f4 br3 b--yellow black-70 mv0 pv2 ph4">Create a label</h1>

<div class="bt b--washed-yellow">
  <form class="pa4 black-80" on:submit|preventDefault={handleSubmit}>
    <div class="measure">
      <input
        class="input-reset ba b--black-20 pa2 mb2 db w-100"
        type="text"
        aria-describedby="title-desc"
        placeholder="Label"
        bind:value={newLabel}
        use:focusThis />
    </div>

    <div class="measure">
      <button type="submit">Submit</button>
    </div>
  </form>
</div>
