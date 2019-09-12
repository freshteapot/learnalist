<svelte:options tag="user-register"/>
<script>
import AlreadyLoggedIn from './already-logged-in.svelte';
import {register, loggedIn} from './shared.js'

let email = '';
let password = '';

function handleSubmit() {
	console.log("How do we want to register users?")
	register(email, password)
}

let isLoggedIn = loggedIn();
if (loggedIn()) {
	console.log('Maybe we just redirect them');
}
</script>

<style>
	input { display: block; width: 500px; max-width: 100%; }
</style>
{#if !isLoggedIn}
<p>Register with email</p>
<form on:submit|preventDefault={handleSubmit}>
	<label>Email</label>
	<input type="text" bind:value={email}/>

	<label>Password</label>
	<input type="password" bind:value={password}/>

	<button disabled={!email || !password} type=submit>
		Register
	</button>
</form>
{:else}
	<AlreadyLoggedIn/>
{/if}
