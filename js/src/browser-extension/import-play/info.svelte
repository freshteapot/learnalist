<script>
  // TODO do I care how often they hit save?
  import { push } from "svelte-spa-router";
  import { loggedIn, notify } from "../../shared.js";
  import store from "./store.js";

  let aList = store.aList;
  let listUrl;
  let show = "overview";
  let saved = false;

  async function handleSave(event) {
    if (saved) {
      return;
    }

    try {
      await store.save();
      listUrl = `${store.getServer()}/alist/${$aList.uuid}.html`;
      show = "saved";
      saved = true;
    } catch (e) {
      saved = false;
      show = "overview";
      notify("error", "Unable to save to learnalist");
    }
  }
</script>

<div class="flex flex-column">
  <div class=" w-100 pa3 mr2">
    <button class="br3" on:click={() => push("/play/total_recall")}>
      Total Recall
    </button>
    <button class="br3" on:click={() => push("/play/slideshow")}>
      Slideshow
    </button>

    <button class="br3" on:click={() => push("/settings")}>Settings</button>
    {#if loggedIn()}
      <button class="br3" on:click={() => push("/spaced_repetition/add")}>
        🧠 + 💪
      </button>

      {#if $aList.info.from.kind != "learnalist"}
        <button class="br3" on:click={handleSave}>Save to Learnalist</button>
      {/if}
    {/if}
  </div>

  <div class="w-100 pa3 mr2">
    {#if show == "overview"}
      <header class="w-100">
        <h1 class="tc">{$aList.info.title}</h1>
      </header>

      <div>
        <table class="w-100" cellspacing="0">
          <thead>
            <tr>
              <th class="fw6 bb b--black-20 pb3 tl">From</th>
              <th class="fw6 bb b--black-20 pb3 tl">To</th>
            </tr>
          </thead>
          <tbody class="lh-copy">
            {#each $aList.data as item, index}
              <tr data-index={index}>
                <td class="pv3 pr3 bb b--black-20">{item.from}</td>
                <td class="pv3 pr3 bb b--black-20">{item.to}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}

    {#if show == "saved"}
      {#if !loggedIn()}
        <p>
          <a target="_blank" href={`${store.getServer()}/login.html`}>
            Log into learnalist.net
          </a>
        </p>
      {:else}
        <p>List has been saved</p>
        <p>
          <a target="_blank" href={listUrl}>Open in the browser</a>
        </p>
      {/if}
    {/if}
  </div>
</div>
