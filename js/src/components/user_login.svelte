<script>
  import { login, notify, api } from "../store.js";
  import { loginHelper } from "../utils/login_helper.js";
  import {
    save as cacheSave,
    KeyUserUuid,
    KeyUserAuthentication
  } from "../cache.js";

  // TODO actually check if logged in
  let isLoggedIn = $loginHelper.loggedIn;
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

    cacheSave(KeyUserUuid, response.body.user_uuid);
    cacheSave(KeyUserAuthentication, response.body.token);
    login("/welcome.html");

    return;
  }
</script>

<style>
  @import "../../all.css";
</style>

<svelte:options tag={null} />

{#if !isLoggedIn}
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
          autocapitalize="none" />
      </div>
      <div class="mv3">
        <label class="db fw6 lh-copy f6" for="password">Password</label>
        <input
          class="b pa2 input-reset ba bg-transparent b--black-20 w-100 br2"
          type="password"
          name="password"
          autocomplete="off"
          bind:value={password}
          id="password" />
      </div>
    </fieldset>
    <div class="measure flex">
      <div class="w-100 items-end">
        <div class="fr">
          <div class="flex items-center mb2">
            <button class="br3 db w-100" type="submit">Login</button>
          </div>
          <div class="flex items-center mb2">
            <span class="f6 link dib black">
              or with
              <a
                target="_blank"
                href="/api/v1/oauth/google/redirect"
                class="f6 link underline dib black">
                google
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
