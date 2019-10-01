<svelte:options tag="v1-menu"/>
<script>
	export let slideshow = "1";
	export let playScreen = "#play";
	export let infoScreen = "#list-info";

	function showSlideshow(event) {
		document.querySelector("v1-slideshow").start(document.querySelector(infoScreen), document.querySelector(playScreen));
	}
</script>
<style>
@import url('/css/tachyons.min.css');
</style>

{#if slideshow == "1"}
<button class="br3" on:click={showSlideshow}>Slideshow</button>
{/if}
