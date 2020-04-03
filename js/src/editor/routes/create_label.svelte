<script>
  import { hasWhiteSpace, focusThis } from "../lib/helper.js";
  import ErrorBox from "../components/error_box.svelte";
  let newLabel = "";
  let message;

  function clearMessage() {
    message = null;
  }

  async function add() {
    if (newLabel === "" || hasWhiteSpace(newLabel)) {
      message = "The label cannot be empty.";
      newLabel = "";
      return;
    }
    let response = await label.save(newLabel);
    if (response.status === 201 || response.status === 200) {
      message = response.body.message;
      router.showScreenLabelView(newLabel);
    } else {
      message = response.body.message;
    }
  }

  async function handleSubmit() {
    if (newLabel === "" || hasWhiteSpace(newLabel)) {
      message = "The label cannot be empty.";
      newLabel = "";
      return;
    }
    /*
  let response = await label.save(newLabel);
  if (response.status === 201 || response.status === 200) {
    message = response.body.message;
    router.showScreenLabelView(newLabel);
  } else {
    message = response.body.message;
  }
  */
  }
</script>

<div class="pv0">
  <section class="ph0 mh0 pv0">
    {#if message}
      <ErrorBox {message} on:clear="{clearMessage}" />
    {/if}

    <article class="mw10 bt bw3 b--yellow mw-100">
      <h1 class="f4 br3 b--yellow mw-100 black-70 mv0 pv2 ph4">
        Create a label
      </h1>

      <div class="bt b--washed-yellow">
        <form class="pa4 black-80" on:submit|preventDefault="{handleSubmit}">
          <div class="measure">
            <input
              class="input-reset ba b--black-20 pa2 mb2 db w-100"
              type="text"
              aria-describedby="title-desc"
              placeholder="Label"
              bind:value="{newLabel}"
              use:focusThis
            />
          </div>

          <div class="measure">
            <button type="submit">Submit</button>
          </div>
        </form>
      </div>
    </article>

  </section>
</div>
