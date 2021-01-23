<svelte:options tag={null} accessors={true} />

<script>
    import { api } from "../../../shared";

    export let alistUuid;
    export let userUuid;
    export let overtimeActive;

    async function stopOvertime(event) {
        console.log("Stop overtime");

        const removed = await api.spacedRepetitionRemoveListFromOvertime(
            userUuid,
            alistUuid
        );

        if (removed) {
            overtimeActive = false;
        }
    }
</script>

<p>
    All items are being added overtime, click to <a
        href="#"
        class="link"
        on:click|once|preventDefault={stopOvertime}>cancel / stop</a
    >
</p>

<style>
    @import "../../../../all.css";
</style>
