<script>
  // TODO do I care how often they hit save?
  // TODO this is not being saved, I suspect due to openapi

  // Could store a list with the relationships
  // {"kind", "cram", "ext_uuid": "setID", "uuid": ""}

  import { push } from "svelte-spa-router";
  import LoginModal from "../../components/login_modal.svelte";
  import { loggedIn, notify } from "../../shared.js";
  import store from "./store.js";

  let aList = $store;
  let listUrl;
  let show = "overview";
  let saved = false;

  async function handleSave(event) {
    if (saved) {
      return;
    }

    try {
      await store.save(aList);
      aList = $store;
      listUrl = `${store.getServer()}/alist/${aList.uuid}.html`;
      show = "saved";
      saved = true;
    } catch (e) {
      console.log("e", e);
      alert("Fail");
    }
  }
</script>

<button class="br3" on:click={() => push('/play/total_recall')}>
  Total Recall
</button>
<button class="br3" on:click={() => push('/play/slideshow')}>Slideshow</button>

<button class="br3" on:click={() => push('/settings')}>Settings</button>
{#if loggedIn()}
  <!--
    - Doesnt work because it needs something below in list-info, I wonder how to solve this
    - maybe build the router manually? as the code underneath should work
  -->
  <button class="br3" on:click={() => push('/interact/spaced_repetition/add')}>
    🧠 + 💪
  </button>

  {#if aList.info.from.kind != 'learnalist'}
    <button class="br3" on:click={handleSave}>Save to Learnalist</button>
  {/if}
{/if}

{#if show == 'overview'}
  <header>
    <h1 class="tc">{aList.info.title}</h1>
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
        {#each aList.data as item, index}
          <tr data-index={index}>
            <td class="pv3 pr3 bb b--black-20">{item.from}</td>
            <td class="pv3 pr3 bb b--black-20">{item.to}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
{/if}

{#if show == 'saved'}
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
