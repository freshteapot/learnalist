<script>
	import { onMount } from "svelte";

	let before = "";
	let rows = [{ from: "", to: "" }];
	let mounted = false;
	let locked = false;
	let sentenceLength = 100;

	function getFromStorage(key, _default) {
		let temp = localStorage.getItem(key);
		return temp ? JSON.parse(temp) : _default;
	}

	onMount(async () => {
		before = getFromStorage("before", "");
		rows = getFromStorage("rows", [{ from: "", to: "" }]);
		locked = getFromStorage("locked", false);
		sentenceLength = getFromStorage("sentenceLength", 100);
		mounted = true;
	});

	$: store("before", before);
	$: store("locked", locked);
	$: store("rows", rows);
	$: store("sentenceLength", sentenceLength);

	function store(key, data) {
		if (!mounted) return;
		localStorage.setItem(key, JSON.stringify(data));
	}

	const insertIntoArray = (arr, value) => {
		return arr.reduce((result, element, index, array) => {
			result.push(element);

			if (index < array.length - 1) {
				const copy = JSON.parse(JSON.stringify(value));
				result.push(copy);
			}

			return result;
		}, []);
	};
</script>

<main>
	<h1>Notepad</h1>
	<button
		class="br3 ma2 pa2"
		on:click={() => {
			rows = [{ from: "", to: "" }];
			locked = false;
			before = "";
		}}>Restart</button
	>

	<!--
	{#if !locked}
		<input
			bind:value={sentenceLength}
			type="range"
			min="40"
			max="100"
			step="10"
		/>
	{/if}
	-->
	{#if locked}
		<button
			class="br3 ma2 pa2"
			on:click={() => {
				console.log("TODO");
			}}>Save</button
		>

		<button
			class="br3 ma2 pa2"
			on:click={() => {
				rows = rows.map((row, index) => {
					return !(index & 1) ? row : { from: "", to: "" };
				});
			}}>Clear</button
		>
	{/if}
	<article class="cf">
		{#if !locked}
			<div class="flex items-center flex-column">
				<div class="outline w-75 pa3 mr2">
					<textarea
						placeholder="Paste some sentences"
						class="pv0 mv0"
						rows="1"
						on:paste={(event) => {
							let paste = (
								event.clipboardData || window.clipboardData
							).getData("text");
							paste = paste.trim();
							store("before", paste);

							paste = paste.replace(/\n\s*\n/g, "\n");

							let parts = paste.split("\n");

							var regex = new RegExp(
								"[\\s\\S]{1," + sentenceLength + "}(?!\\S)",
								"g"
							);

							let parts2 = parts.flatMap((e) => {
								let parts = e
									//.replace(/[\s\S]{1,100}(?!\S)/g, "$&\n")
									.replace(regex, "$&\n")
									.split("\n");
								return parts.map((e) => {
									return e.trimStart();
								});
							});

							rows = parts2
								.map((e) => {
									return { from: e, to: "" };
								})
								.filter((e) => !(e.from === "" && e.to === ""));

							rows = [
								...insertIntoArray(rows, {
									from: "",
									to: "",
								}),
								{ from: "", to: "" },
							];
							locked = true;

							store("rows", rows);
							store("locked", locked);
						}}
					/>
				</div>
			</div>{/if}

		{#if locked}
			<div class="flex items-center flex-column">
				<div class="outline w-75 pa3 mr2">
					{#each rows as row, index}
						<textarea
							class="pv0 mv0"
							learn
							rows="1"
							disabled={!(index & 1)}
							class:learn={index & 1}
							bind:value={row.from}
							bind:this={row.elFrom}
							on:paste={(event) => {
								event.preventDefault();
								return false;
							}}
						/>
					{/each}
				</div>
			</div>
		{/if}
	</article>
</main>

<style>
	@import "../../../all.css";
	textarea {
		border: 1px solid #eeeeee;
		box-shadow: 1px 1px 0 #dddddd;
		display: block;
		font-size: 22px;
		line-height: 50px;

		resize: none;
		height: 100%;
		width: 100%;

		background-image: -moz-linear-gradient(
			top,
			transparent,
			transparent 49px,
			#e7eff8 0px
		);
		background-image: -webkit-linear-gradient(
			top,
			transparent,
			transparent 49px,
			#e7eff8 0
		);

		-webkit-background-size: 100% 50px;
		background-size: 100% 50px;
	}

	/* */
	main {
		text-align: center;
		margin: 0 auto;
	}

	textarea:focus {
		outline: none !important;
	}

	textarea:read-only {
		border: 0;
		color: #000;
		box-shadow: none;
		background-color: white;
	}

	.learn {
		color: rgb(37, 47, 63);
		background-color: rgb(252, 233, 106);
	}
</style>
