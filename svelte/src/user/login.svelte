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
@import url('/css/tachyons.min.css');
</style>
<main class="pa4 black-80">
{#if !isLoggedIn}
	<form class="measure center" on:submit|preventDefault={handleSubmit}>
		<fieldset id="sign_up" class="ba b--transparent ph0 mh0">
			<div class="mt3">
				<label class="db fw6 lh-copy f6" for="email-address">Email</label>
				<input class="pa2 input-reset ba bg-transparent b--black-20 w-100 br2" type="email" name="email-address" bind:value={email} id="email-address">
			</div>
			<div class="mv3">
				<label class="db fw6 lh-copy f6" for="password">Password</label>
				<input class="b pa2 input-reset ba bg-transparent b--black-20 w-100 br2" type="password" name="password" bind:value={password} id="password">
			</div>
		</fieldset>
		<div class="">
			<input class="b ph3 pv2 input-reset ba b--black bg-transparent grow pointer f6 dib br2" disabled={!email || !password} type="submit" value="L">
			<input class="b ph3 pv2 input-reset ba b--black bg-transparent grow pointer f6 dib br2" disabled={!email || !password} type="submit" value="O">
			<input class="b ph3 pv2 input-reset ba b--black bg-transparent grow pointer f6 dib br2" disabled={!email || !password} type="submit" value="G">
			<input class="b ph3 pv2 input-reset ba b--black bg-transparent grow pointer f6 dib br2" disabled={!email || !password} type="submit" value="I">
			<input class="b ph3 pv2 input-reset ba b--black bg-transparent grow pointer f6 dib br2" disabled={!email || !password} type="submit" value="N">
		</div>


		<div class="lh-copy mt3">
			<a href="/forgot-your-password.html" class="f6 link dim black db">Forgot your password?</a>
			<a href="/register.html" class="f6 link dim black db">Sign up</a>
		</div>
	</form>

{:else}
<p class="measure center">
You are already logged in.<br/>
Goto the <a href='/welcome.html'>welcome page</a>
</p>
{/if}
</main>
