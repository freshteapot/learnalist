import App from './App.svelte';

var app = new App({
	props: {
		aList: {
			uuid: "",
			data: [],
			info: {
				title: ""
			}
		}
	}
});

export default app;
