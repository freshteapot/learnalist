<svelte:options tag="v1-menu"/>
<script>
	export let slideshow = "1";
	export let playScreen = "#play";
	export let infoScreen = "#list-info";

	function showSlideshow(event) {
		document.querySelector(playScreen).style.display = "";
		document.querySelector(infoScreen).style.display = "none";
	}
</script>
{#if slideshow == "1"}
<button on:click={showSlideshow}>Slideshow</button>
{/if}
