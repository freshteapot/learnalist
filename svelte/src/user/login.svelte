<svelte:options tag='user-login'/>
<script>
import AlreadyLoggedIn from './already-logged-in.svelte';
import {logIn, loggedIn} from './shared.js'

export let redirectOnLogin='/welcome.html';

let email = '';
let password = '';

function handleSubmit() {
	console.log('Try logging them in');
	console.log('email is ' + email);
	console.log('password is ' + password);

	// TODO success
	console.log('// TODO: Get the token from the server');
	logIn('abc123', redirectOnLogin)
}

let isLoggedIn = loggedIn();
</script>

<style>
	input { display: block; width: 500px; max-width: 100%; }
</style>

{#if !isLoggedIn}
<p>Login with</p>
<form on:submit|preventDefault={handleSubmit}>
	<label>Email</label>
	<input type='text' bind:value={email} />

	<label>Password</label>
	<input type='password' bind:value={password} />

	<button disabled={!email || !password} type=submit>
		Login
	</button>
	<p>
		or <a href="/register.html">register</a>
	</p>
</form>
{:else}
<AlreadyLoggedIn/>
{/if}
