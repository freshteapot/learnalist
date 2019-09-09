<svelte:options tag="v1-slideshow"/>
<script>
	export let aList = {
		uuid: "",
		data: [""],
		info: {
			title: ""
		}
	}
	export let playScreen = "#play"
	export let infoScreen = "#list-info"

	let loops = 0;
	let index = 0;
	let show = "Welcome, to begin, click next.";
	let nextTimeIsLoop = 0;

	function handleClick(event) {
		if (nextTimeIsLoop) {
			loops += 1;
			nextTimeIsLoop = 0;
		}

		show = aList.data[index];
		index += 1;
		if (!aList.data[index]) {
			index = 0;
			nextTimeIsLoop = 1;
		}
	}

	function handleClose(event) {
		document.querySelector(playScreen).style.display = "none";
		document.querySelector(infoScreen).style.display = ""
		loops = 0;
		index = 0;
		nextTimeIsLoop = 0;
		show  = "";
	}
</script>
<button on:click={handleClose}>Close</button>
<h1>Slideshow</h1>
<button on:click={handleClick}>Next</button>
<p>{show}</p>
{#if loops > 0}
<p>How many times have you clicked thru the list?</p>
<p>{loops}</p>
{/if}
