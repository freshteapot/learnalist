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

<button class="br3" on:click={() => push('/start')}>Close</button>

<div>
  <h1>Change server</h1>
  <p>
    You only need to change this if you are running your own learnalist server
    or developing the chrome extension
  </p>
  <p>
    <input bind:value={baseUrl} />
  </p>
  <button class="br3" on:click|preventDefault={handleSubmit}>Submit</button>
</div>

<div>
  <h1>Reset to default settings</h1>
  <p>
    <button class="br3" on:click|preventDefault={handleReset}>Reset</button>
  </p>
</div>
