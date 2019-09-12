<svelte:options tag="user-login"/>
<script>
export let lastViewedKey="last.viewed";
export let redirectOnLogin="/welcome.html";

let email = '';
let password = '';

function handleSubmit() {
	console.log("Try logging them in")
	console.log("email is " + email)
	console.log("password is " + password)

	let lastViewed = localStorage.getItem(lastViewedKey);
	if (!lastViewed) {
		console.log(redirectOnLogin)
		lastViewed = redirectOnLogin;
	}

	window.location = lastViewed;
}
</script>

<style>
	input { display: block; width: 500px; max-width: 100%; }
</style>
<p>Login with</p>
<form on:submit|preventDefault={handleSubmit}>
	<label>Email</label>
	<input type="text" bind:value={email} />

	<label>Password</label>
	<input type="password" bind:value={password} />

	<button disabled={!email || !password} type=submit>
		Login
	</button>
</form>
