<script>
  import { push } from "svelte-spa-router";
  import Header from "./header.svelte";
  import Info from "./info.svelte";
  import store from "./store.js";
  import { onMount } from "svelte";
  // Add new
  // Add if block
  // Add to url.inclues

  let aList;
  let show = "";
  let assumeFailedToFindList = null;

  onMount(async () => {
    listenForMessagesFromBrowser();
    checkForLists();
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

  function listenForMessagesFromBrowser() {
    chrome.runtime.onMessageExternal.addListener(function(
      request,
      sender,
      sendResponse
    ) {
      try {
        clearTimeout(assumeFailedToFindList);
        assumeFailedToFindList = null;
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
        if (!aList) {
          throw "list.not.found";
        }
        store.load(aList);
        push("/overview");
      } catch (e) {
        show = "not-supported";
      }
    });
  }

  function checkForLists() {
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
      assumeFailedToFindList = setTimeout(() => {
        show = "welcome";
      }, 100);
    });
  }
</script>

<Header />
{#if show != ''}
  <div class="flex flex-column">
    {#if show == 'welcome'}
      <div class="w-100 pa3 mr2">
        <h1>Welcome!</h1>
        <p>We will only try and load lists from</p>
        <ul class="list">
          <li>
            <a href="https://quizlet.com" target="_blank">quizlet.com</a>
          </li>
          <li>
            <a href="https://cram.com" target="_blank">cram.com</a>
          </li>
          <li>
            <a href="https://brainscape.com" target="_blank">brainscape.com</a>
          </li>
          <li>
            <a href="https://learnalist.net" target="_blank">learnalist.net</a>
          </li>
        </ul>
      </div>
    {/if}

    {#if show == 'not-supported'}
      <div class=" w-100 pa3 mr2">
        <p>We were unable to find a list on this page.</p>
        <p>Do you think this is a bug? Let us know</p>
      </div>
    {/if}
  </div>
{/if}
