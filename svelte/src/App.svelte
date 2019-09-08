<svelte:options tag="my-thing"/>
<script>
	export let aList = {
		uuid: "",
		data: [],
		info: {
			title: ""
		}
	};

</script>
<h1>Title is {aList.info.title}!</h1>
