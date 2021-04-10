<script>
	import stripeData from "./stripe.json";
	import LoginModal from "../../components/login_modal.svelte";
	import { loggedIn } from "../shared.js";

	let stripe;

	const loginNagMessageDefault =
		"You need to be logged in so we can link your payment to your user.";
	let loginNagMessage = loginNagMessageDefault;
	let loginNagShown = false;

	let options = stripeData.options;
	let picked = "";
	let currentCurrency = "";

	function stripeLoaded() {
		stripe = Stripe(stripeData.key);
	}

	function breadwinner() {
		if (!loggedIn()) {
			loginNagShown = false;
			return;
		}

		const data = {
			price_id: picked,
		};

		try {
			// TODO payment or payments?
			fetch("/payment/create-checkout-session", {
				method: "POST",
				body: JSON.stringify(data),
			})
				.then(function (response) {
					return response.json();
				})
				.then(function (session) {
					return stripe.redirectToCheckout({ sessionId: session.id });
				})
				.then(function (result) {
					// If redirectToCheckout fails due to a browser or network
					// error, you should display the localized error message to your
					// customer using error.message.
					if (result.error) {
						alert(result.error.message);
					}
				})
				.catch(function (error) {
					console.error("Error:", error);
					alert(
						"Sadly the internet doesn't want to let us take your money :("
					);
				});
		} catch (e) {
			// Assume fetch not supported
			alert(
				"We cant take your money due to your browser not supporting fetch."
			);
		}
	}

	function closeLoginModal() {
		loginNagShown = true;
		return;
	}

	$: readyToPay = stripe && currentCurrency !== "" && picked !== "";
	$: currencies = [
		"",
		...new Set(options.map((e) => e.currency.toUpperCase())),
	];
	$: prices = options.filter(
		(e) => e.currency === currentCurrency.toLowerCase()
	);
</script>

<div class="pa2">
	<h2>1) Pick a Currency</h2>
	<!-- svelte-ignore a11y-no-onchange -->
	<select bind:value={currentCurrency}>
		{#each currencies as currency}
			<option value={currency}>
				{currency}
			</option>
		{/each}
	</select>

	{#if currentCurrency}
		<h2>2) Pick amount</h2>
		{#each prices as price}
			<label>
				<input type="radio" bind:group={picked} value={price.id} />
				{price.human_amount} ({price.currency.toUpperCase()})
			</label>
		{/each}
	{/if}

	{#if readyToPay}
		<h2>3) Support</h2>
		<button class="br3" on:click={breadwinner}>Go to checkout</button>
	{/if}

	{#if !loggedIn() && !loginNagShown}
		<LoginModal on:close={closeLoginModal}>
			<p>{loginNagMessage}</p>
		</LoginModal>
	{/if}
</div>

<svelte:head>
	<!--
	<script
		src="https://polyfill.io/v3/polyfill.min.js?version=3.52.1&features=fetch" ✂prettier:content✂="" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=" ✂prettier:content✂="e30=">{}</script>
	-->
	<script src="https://js.stripe.com/v3/" on:load={stripeLoaded}></script>
</svelte:head>
