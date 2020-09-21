<script>
  import { push } from "svelte-spa-router";

  import {
    saveConfiguration,
    getConfiguration,
    clearConfiguration,
    KeySettingsServer
  } from "../../configuration.js";

  let baseUrl = getConfiguration(KeySettingsServer, "https://learnalist.net");

  function handleSubmit() {
    clearConfiguration();
    saveConfiguration(KeySettingsServer, baseUrl);
    chrome.runtime.sendMessage({ kind: "lookup-login-info" });
  }

  function handleReset() {
    clearConfiguration();
    baseUrl = getConfiguration(KeySettingsServer, "https://learnalist.net");
    chrome.runtime.sendMessage({ kind: "lookup-login-info" });
  }
</script>

<div class="flex flex-column">
  <div class=" w-100 pa3 mr2">
    <h1 class="f2 measure">Settings</h1>
    <button class="br3" on:click={() => push('/start')}>Close</button>
  </div>

  <div class="w-100 pa3 mr2">
    <h2>Change server</h2>
    <p>
      You only need to change this if you are running your own learnalist server
      or developing the chrome extension
    </p>
    <p>
      <input class="w-100 pa3 mr2" bind:value={baseUrl} />
    </p>
    <button class="br3" on:click|preventDefault={handleSubmit}>Submit</button>
  </div>

  <div class=" w-100 pa3 mr2">
    <h2>Reset to default settings</h2>
    <p>
      <button class="br3" on:click|preventDefault={handleReset}>Reset</button>
    </p>
  </div>
</div>
