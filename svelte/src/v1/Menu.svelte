<svelte:options tag="v1-menu"/>
<script>
	export let playScreen = "#play"
	export let infoScreen = "#list-info"
	console.log(playScreen)
	function showSlideshow(event) {
		console.log(playScreen)
		console.log(document.querySelector(playScreen))
		document.querySelector(playScreen).style.display = "";
		document.querySelector(infoScreen).style.display = "none"
	}

</script>
<button on:click={showSlideshow}>Slideshow</button>
