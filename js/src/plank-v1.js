
import HumblePlank from "./plank/index.svelte";
import { Configuration, DefaultApi } from "./openapi";

import { getApiServer } from "./utils/setup.js";
import { getListsByMe, getPlanks } from "./api2.js";

var config = new Configuration({
    basePath: `${getApiServer()}/api/v1`
});

var api = new DefaultApi(config);


api.getServerVersion().then(function (data) {
    console.log('API called successfully. Returned data: ' + data);
    console.log(data);
}, function (error) {
    console.error(error);
});

getListsByMe().then(data => {
    console.log("all my lists", data);
})

getPlanks().then(data => {
    console.log("planks", data);
})

// Actual app to handle the interactions
let app;
const el = document.querySelector("#list-info")
if (el) {
    app = new HumblePlank({
        target: el,
    });
}

export default app;
