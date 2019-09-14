<svelte:options tag="user-register"/>
<script>
import AlreadyLoggedIn from './already-logged-in.svelte';
import {register, loggedIn} from './shared.js'

let email = '';
let password = '';
let usersname = '';
function handleSubmit() {
	console.log('How do we want to register users?')
	register(email, password)
	console.log('Users name is ' + usersname);
}

let isLoggedIn = loggedIn();
if (loggedIn()) {
	console.log('Maybe we just redirect them');
}
</script>

<style>
@import url('/css/tachyons.min.css');
</style>
{#if !isLoggedIn}

<div class="pa4-l">
	<form class="bg-yellow mw7 center pa4 br2-ns ba b--black-10" on:submit|preventDefault={handleSubmit}>
		<fieldset class="cf bn ma0 pa0">
			<legend class="pa0 f5 f4-ns mb3 black-80">Register and start making lists</legend>
			<div class="mt3">
				<label class="db fw6 lh-copy f6" for="email-address">Email</label>
				<input class="pa2 input-reset ba bg-white b--black-20 w-100 br2" type="email" name="email-address" bind:value={email}  id="email">
			</div>
			<div class="mv3">
				<label class="db fw6 lh-copy f6" for="password">Password</label>
				<input class="b pa2 input-reset ba bg-white b--black-20 w-100 br2" type="password" name="password" bind:value={password} id="password">
			</div>
			<div class="mv3">
				<label class="db fw6 lh-copy f6" for="usersname">Name <span class="normal black-60">(optional)</span></label>
				<input class="b pa2 input-reset ba bg-white b--black-20 w-100 br2" type="text" name="usersname" bind:value={usersname} id="usersname">
			</div>
		</fieldset>
		<div class="mv3">
		<input class="b pa2 input-reset ba bg-white b--black-20 w-100 br2" disabled={!email || !password} type="submit" value="Lets go!">
		</div>
	</form>
</div>
{:else}
	<AlreadyLoggedIn/>
{/if}
