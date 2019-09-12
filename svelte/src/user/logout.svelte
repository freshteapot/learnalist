<svelte:options tag='user-logout'/>
<script>
	import {loggedIn, logOut} from './shared.js';
	export let redirect = '/';

function handleSubmit() {
	logOut(redirect);
}
</script>

<style>
	input { display: block; width: 500px; max-width: 100%; }
</style>

{#if loggedIn()}
<form on:submit|preventDefault={handleSubmit}>
	<button type=submit>
		Logout
	</button>
</form>
{/if}
