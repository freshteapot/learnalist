import { push } from "svelte-spa-router";

const intro = () => push("/intro");
const overview = () => push("/overview");
const add = () => push("/add");
const remind = () => push("/remind");

export default {
    intro,
    overview,
    add,
    remind
}
