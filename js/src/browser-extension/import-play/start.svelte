<script>
  import { push } from "svelte-spa-router";
  import Info from "./info.svelte";
  import store from "./store.js";
  import { onMount } from "svelte";
  // Add new
  // Add if block
  // Add to url.inclues

  let aList;
  let show = "";

  onMount(async () => {
    handle();
  });

  function brainscapeToAlist(input) {
    const data = input.detail;
    return {
      info: {
        title: data.title,
        type: "v2"
      },
      data: data.listData
    };
  }

  function cramToAlist(input) {
    // setID = UUID
    const data = input.detail;
    const listData = data.listData.map(term => {
      return { from: term.front_plain, to: term.back_plain };
    });

    return {
      info: {
        title: data.title,
        type: "v2"
      },
      data: listData
    };
  }

  function quizletToAlist(input) {
    const data = input.detail;
    const listData = Object.values(
      data.listData.setPageData.termIdToTermsMap
    ).map(term => {
      return { from: term.word, to: term.definition };
    });

    return {
      info: {
        title: data.title,
        type: "v2"
      },
      data: listData
    };
  }

  function handle(event) {
    chrome.runtime.onMessageExternal.addListener(function(
      request,
      sender,
      sendResponse
    ) {
      if (request.kind == "quizlet") {
        aList = quizletToAlist(request);
      }

      if (request.kind == "cram") {
        aList = cramToAlist(request);
      }

      if (request.kind == "brainscape") {
        aList = brainscapeToAlist(request);
      }

      aList = aList;

      document.querySelector("#play-data").innerHTML = JSON.stringify(aList);

      if (aList) {
        store.load(aList);
        push("/overview");
      }
    });

    chrome.tabs.query({ active: true, currentWindow: true }, function(tabs) {
      const load =
        tabs[0].url.includes("cram.com") ||
        tabs[0].url.includes("quizlet.com") ||
        tabs[0].url.includes("brainscape.com");

      if (!load) {
        show = "welcome";
        return;
      }

      show = "";
      chrome.tabs.sendMessage(tabs[0].id, { greeting: "hello" });
    });
  }

  $: console.log("show", show);
</script>

<style>
  main {
    text-align: center;
    padding: 1em;
    width: 500px;
    margin: 0 auto;
  }
</style>

{#if show == 'welcome'}
  <button class="br3" on:click={() => push('/settings')}>Settings</button>
  <main>
    <h1>Welcome!</h1>
    <p>We will only try and load the list from</p>
    <ul class="list">
      <li>quizlet.com</li>
      <li>cram.com</li>
      <li>brainscape.com</li>
    </ul>

  </main>
{/if}
