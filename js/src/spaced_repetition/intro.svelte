<script>
  import { location } from "svelte-spa-router";
  import goto from "./goto.js";
  import { loggedIn } from "../shared.js";

  let listElement = document.querySelector("#list-info");
  let playElement = document.querySelector("#play");

  if (window.location.search) {
    window.location = `${window.location.origin}/${window.location.pathname}${window.location.hash}`;
  }

  if ($location === "/intro") {
    listElement.style.display = "";
    playElement.style.display = "none";
  } else {
    if (loggedIn()) {
      goto.overview();
    } else {
      listElement.style.display = "";
      playElement.style.display = "none";
    }
  }
</script>
