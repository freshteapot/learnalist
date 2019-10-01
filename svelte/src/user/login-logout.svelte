<svelte:options tag='user-logout'/>
<script>
	import {loggedIn, logOut} from './shared.js';
	export let redirect = '/';

function handleSubmit() {
	logOut(redirect);
}
let isLoggedIn = loggedIn();
</script>

<style>
@import url('/css/tachyons.min.css');
</style>

{#if isLoggedIn}
<a title="Login or logout" href="/login.html" on:click|once={handleSubmit} class="f6 fw6 hover-black link black-70 mr2 mr3-m mr4-l dib-l">
	Logout
</a>
{:else}
<a title="Login or logout" href="/login.html" class="f6 fw6 hover-black link black-70 mr2 mr3-m mr4-l dib-l">
	Login
</a>
{/if}
