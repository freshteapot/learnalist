<script>
  import { push } from "svelte-spa-router";
  import { loginHelper } from "../lib/helper.js";
  import { postLogin } from "../lib/api.js";
  import ErrorBox from "./error_box.svelte";

  let isLoggedIn = $loginHelper.loggedIn;
  let username = "";
  let password = "";
  let message;

  async function handleSubmit() {
    if (username === "" || password === "") {
      message = "Please enter in a username and password";
      return;
    }

    let response = await postLogin(username, password);
    if (response.status != 200) {
      alert("Try again");
      return;
    }

    loginHelper.login(response.body);
    push($loginHelper.redirectURL);
    // loginHelper.redirectURLAfterLogin();
    return;
  }

  function clearMessage() {
    message = null;
  }
</script>

{#if message}
  <ErrorBox {message} on:clear="{clearMessage}" />
{/if}
<main class="pa4 black-80">
  {#if !isLoggedIn}
    <form class="measure center" on:submit|preventDefault="{handleSubmit}">

      <fieldset id="sign_up" class="ba b--transparent ph0 mh0">
        <div class="mt3">
          <label class="db fw6 lh-copy f6" for="username">Username</label>
          <input
            class="pa2 input-reset ba bg-transparent b--black-20 w-100 br2"
            type="text"
            name="username"
            bind:value="{username}"
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
            bind:value="{password}"
            id="password"
          />
        </div>
      </fieldset>

      <div class="measure flex">
        <div class="w-100 items-end">
          <div class="fr">
            <div class="flex items-center mb2">
              <button class="db w-100" type="submit">Login</button>
            </div>

            <div class="flex items-center mb2">
              <span class="f6 link dib black">
                or with
                <a
                  target="_blank"
                  href="https://learnalist.net/api/v1/oauth/google/redirect"
                  class="f6 link underline dib black"
                >
                  google
                </a>
              </span>
            </div>

            <div class="flex items-center mb2">
              <span class="f6 link dib black">
                or via
                <a
                  target="_blank"
                  href="https://learnalist.net/login.html"
                  class="f6 link underline dib black"
                >
                  learnalist login
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
</main>
