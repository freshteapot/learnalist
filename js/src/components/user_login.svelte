<svelte:options tag={null} />

<script>
  import { login, notify, api } from "../shared.js";
  import { loginHelper } from "../utils/login_helper.js";
  import {
    saveConfiguration,
    KeyUserUuid,
    KeyUserAuthentication,
  } from "../configuration.js";

  let username = "";
  let password = "";
  let message;

  // https://github.com/sveltejs/svelte/issues/2118#issuecomment-531586875
  async function handleSubmit() {
    if (username === "" || password === "") {
      message = "Please enter in a username and password";
      notify("error", message);
      return;
    }

    const response = await api.postLogin(username, password);

    if (response.status != 200) {
      notify("error", "Please try again");
      return;
    }

    saveConfiguration(KeyUserUuid, response.body.user_uuid);
    saveConfiguration(KeyUserAuthentication, response.body.token);

    const querystring = window.location.search;
    const searchParams = new URLSearchParams(querystring);

    if (!searchParams.has("redirect")) {
      login("/welcome.html");
      return;
    }

    const suffix = searchParams.get("redirect").replace(/^\/+/, "");
    const redirectUrl = addLoginRedirect(`${api.getServer()}/${suffix}`);

    login(redirectUrl);
    return;
  }

  function addLoginRedirect(redirectUrl) {
    const url = document.createElement("a");
    url.href = redirectUrl;
    const querystring = url.search;
    const searchParams = new URLSearchParams(querystring);
    searchParams.set("login_redirect", "true");
    url.search = `?${searchParams.toString()}`;
    return url.href.replace(url.origin, "");
  }
</script>

{#if !$loginHelper.loggedIn}
  <form class="measure center" on:submit|preventDefault={handleSubmit}>
    <fieldset id="sign_up" class="ba b--transparent ph0 mh0">
      <div class="mt3">
        <label class="db fw6 lh-copy f6" for="username">Username</label>
        <input
          class="pa2 input-reset ba bg-transparent b--black-20 w-100 br2"
          type="text"
          name="username"
          bind:value={username}
          id="username"
          autocapitalize="none"
        />
      </div>
      <div class="mv3">
        <label class="db fw6 lh-copy f6" for="password">Password</label>
        <input
          class="b pa2 input-reset ba bg-transparent b--black-20 w-100 br2"
          type="password"
          name="password"
          autocomplete="off"
          bind:value={password}
          id="password"
        />
      </div>
    </fieldset>
    <div class="measure flex">
      <div class="w-100 items-end">
        <div class="fr">
          <div class="flex items-center mb2">
            <button class="br3 db w-100" type="submit">Login</button>
          </div>
          <div class="flex items-center mb2">
            <span class="f6 link dib black"> or with </span>
          </div>

          <div class="flex items-center mb2">
            <a
              target="_blank"
              href="/api/v1/oauth/google/redirect"
              class="f6 link underline dib black"
            >
              google
            </a>
          </div>
          <div class="flex items-center mb2">
            <span class="f6 link dib black">
              <a
                target="_blank"
                href="/api/v1/oauth/appleid/redirect"
                class="f6 link underline dib black"
              >
                apple
              </a>
            </span>
          </div>
        </div>
      </div>
    </div>
  </form>
{:else}
  <p class="measure center">
    You are already logged in.
    <br />
    Goto the
    <a href="/welcome.html">welcome page</a>
  </p>
{/if}

<style>
  @import "../../all.css";
</style>
