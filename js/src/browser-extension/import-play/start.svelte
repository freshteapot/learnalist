<script>
  import { push } from "svelte-spa-router";
  import store from "./store.js";
  import { onMount } from "svelte";
  import cram from "./providers/cram.js";
  import brainscape from "./providers/brainscape.js";
  import quizlet from "./providers/quizlet.js";
  import learnalist from "./providers/learnalist.js";
  import { getConfiguration } from "../../configuration.js";
  import { clearNotification } from "../../shared.js";

  // Add new
  // Add if block
  // Add to url.inclues

  let aList;
  let show = "";
  let assumeFailedToFindList = null;
  const providers = [learnalist, quizlet, brainscape, cram];
  const mappers = Object.fromEntries(providers.map((e) => [e.key, e.convert]));
  const domains = providers.map((e) => e.domain);

  onMount(async () => {
    clearNotification();
    // Development feature
    const localDomain = getConfiguration("dev.checklist.domain", "");
    if (localDomain != "") {
      domains.push(localDomain);
    }

    listenForMessagesFromBrowser(mappers);
    checkForLists(domains);
  });

  function listenForMessagesFromBrowser(mappers) {
    chrome.runtime.onMessageExternal.addListener(function (request) {
      try {
        clearTimeout(assumeFailedToFindList);
        assumeFailedToFindList = null;
        // Mapping based on kind
        if (mappers.hasOwnProperty(request.kind)) {
          const mapper = mappers[request.kind];
          aList = mapper(request);
        }

        // TODO do i trim v2 in the data?
        // Trim entries
        aList.data.map((entry) => {
          entry.from = entry.from.trim();
          entry.to = entry.to.trim();
          return entry;
        });

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

  function checkForLists(allowedDomains) {
    chrome.tabs.query({ active: true, currentWindow: true }, function (tabs) {
      const load = allowedDomains.some((domain) =>
        tabs[0].url.includes(domain)
      );

      if (!load) {
        show = "welcome";
        return;
      }

      show = "";
      // Part of debugging
      chrome.tabs.sendMessage(tabs[0].id, { kind: "load-data" });
      assumeFailedToFindList = setTimeout(() => {
        show = "welcome";
      }, 100);
    });
  }
</script>

{#if show != ""}
  <div class="flex flex-column">
    <div class="w-100 pa3 mr2">
      <button class="br3" on:click={() => push("/settings")}>Settings</button>
    </div>
    {#if show == "welcome"}
      <div class="w-100 pa3 mr2">
        <h1>Welcome!!</h1>
        <p>We will only try and load lists from</p>
        <ul class="list">
          {#each providers as provider}
            <li>
              <a href={provider.url} target="_blank">{provider.domain}</a>
            </li>
          {/each}
        </ul>
      </div>
    {/if}

    {#if show == "not-supported"}
      <div class="w-100 pa3 mr2">
        <p>We were unable to find a list on this page.</p>
        <p>Do you think this is a bug? Let us know</p>
      </div>
    {/if}
  </div>
{/if}
