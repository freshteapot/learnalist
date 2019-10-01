<svelte:options tag="user-menu"/>
<script>
import LoginLogout from './login-logout.svelte';
import {loggedIn} from './shared.js'

let isLoggedIn = loggedIn();
</script>

<style>
@import url('/css/tachyons.min.css');
</style>
<div class="db dt-ns mw9 center w-100">
	<nav class="pa3 pa1-ns">
		<a class="link dim black b f6 f5-ns dib mr3" href="/" title="Home">Learnalist</a>
	</nav>
  <nav class="db dtc-ns v-mid w-100 tl tr-ns mt2 mt0-ns">
{#if isLoggedIn}
		<a title="Create your own list" href="/create-your-own-list.html" class="f6 fw6 hover-black link black-70 mr2 mr3-m mr4-l dib-l">
			Create
		</a>
		<a title="My lists" href="/alists/overview.html" class="f6 fw6 hover-black link black-70 mr2 mr3-m mr4-l dib-l">
			My Lists
		</a>
		<user-logout></user-logout>
{:else}
		<a title="Login or logout" href="/login.html" class="f6 fw6 hover-black link black-70 mr2 mr3-m mr4-l dib-l">
			Login
		</a>
{/if}
</nav>
</div>
