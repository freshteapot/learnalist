<svelte:options tag={null} accessors={true} />

<script>
    import { removeListFromOvertime } from "../../../spaced_repetition/api.js";
    export let alistUuid;
    export let userUuid;
    export let overtimeActive;

    async function stopOvertime() {
        const removed = await removeListFromOvertime(userUuid, alistUuid);
        if (removed) {
            overtimeActive = false;
        }
    }
</script>

<p>
    All items are being added over time, click to <a
        href="#"
        class="link"
        on:click|once|preventDefault={stopOvertime}>cancel / stop</a
    >
</p>

<style>
    @import "../../../../all.css";
</style>
