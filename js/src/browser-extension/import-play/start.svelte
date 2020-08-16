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

  function learnalistToAlist(input) {
    const aList = input.detail;
    if (aList.info.type !== "v2") {
      throw "Not v2";
    }
    aList.info.from = input.metadata;
    return aList;
  }

  function brainscapeToAlist(input) {
    const data = input.detail;
    return {
      info: {
        title: data.title,
        type: "v2",
        from: input.metadata
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
        type: "v2",
        from: input.metadata
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
        type: "v2",
        from: input.metadata
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
      try {
        if (request.kind == "quizlet") {
          aList = quizletToAlist(request);
        }

        if (request.kind == "cram") {
          aList = cramToAlist(request);
        }

        if (request.kind == "brainscape") {
          aList = brainscapeToAlist(request);
        }

        if (request.kind == "learnalist") {
          aList = learnalistToAlist(request);
        }

        aList = aList;

        document.querySelector("#play-data").innerHTML = JSON.stringify(aList);

        if (aList) {
          store.load(aList);
          push("/overview");
        }
      } catch (e) {
        show = "not-supported";
      }
    });

    chrome.tabs.query({ active: true, currentWindow: true }, function(tabs) {
      const load =
        tabs[0].url.includes("cram.com") ||
        tabs[0].url.includes("quizlet.com") ||
        tabs[0].url.includes("brainscape.com") ||
        tabs[0].url.includes("learnalist.net") ||
        tabs[0].url.includes("localhost:1234");

      if (!load) {
        show = "welcome";
        return;
      }

      show = "";
      chrome.tabs.sendMessage(tabs[0].id, { greeting: "hello" });
    });
  }
</script>

<button class="br3" on:click={() => push('/settings')}>Settings</button>

{#if show == 'welcome'}
  <h1>Welcome!</h1>
  <p>We will only try and load from</p>
  <ul class="list">
    <li>quizlet.com</li>
    <li>cram.com</li>
    <li>brainscape.com</li>
    <li>learnalist.net</li>
  </ul>
{/if}

{#if show == 'not-supported'}
  <p>Content on the page not supported</p>
{/if}
