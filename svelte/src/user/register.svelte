<svelte:options tag="user-register"/>
<script>
export let displayif="/register.html"
let email = '';

function handleSubmit() {
	console.log("How do we want to register users?")
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
<p>Register with email</p>
<form on:submit|preventDefault={handleSubmit}>
	<label>Email</label>
	<input type="text" bind:value={email}/>

	<button disabled={!email} type=submit>
		Register
	</button>
</form>
{/if}
