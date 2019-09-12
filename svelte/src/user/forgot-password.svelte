<svelte:options tag="user-forgot-password"/>
<script>
	export let displayif="/forgotten-password.html"
	let email = '';

function handleSubmit() {
	console.log("Try sending them the password")
	console.log("email is " + email)
}

function showPageIf() {
	if (window.location.pathname === displayif) {
		return true
	}
	return false
}

</script>

<style>
	input { display: block; width: 500px; max-width: 100%; }
</style>
{#if showPageIf()}
<p>Forgotten password</p>
<form on:submit|preventDefault={handleSubmit}>
	<label>Email</label>
	<input type="text" bind:value={email} />

	<button disabled={!email} type=submit>
		Send password
	</button>
</form>
{/if}
