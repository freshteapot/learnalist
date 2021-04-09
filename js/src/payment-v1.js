import Payment from "./payment/v1/index.svelte";
let app;
const el = document.querySelector("#payment-v1-panel")
if (el) {
    app = new Payment({
        target: el,
    });
}

export default app;
