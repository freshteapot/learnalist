<svelte:options tag="user-logout"/>
<script>
	export let displayWhenNot = ["/login.html", "/register.html", "/forgotten-password.html"];
	export let redirect = "/"

function handleSubmit() {
	console.log("Clear cookies")
	console.log("Clear localstorage")
	window.localStorage.clear();
	window.location = redirect
}

function showPageIf() {
	let lookUp = window.location.pathname
	if (displayWhenNot.indexOf(lookUp) !== -1) {
		return false;
	}

	return true
}

</script>

<style>
	input { display: block; width: 500px; max-width: 100%; }
</style>

{#if showPageIf()}
<form on:submit|preventDefault={handleSubmit}>
	<button type=submit>
		Logout
	</button>
</form>
{/if}
