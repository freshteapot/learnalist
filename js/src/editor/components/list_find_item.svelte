<script>
  import goto from "../lib/goto.js";

  export let title = "";
  export let uuid = "";
  function handleClick() {
    console.log(uuid);
    goto.list.view(uuid);
  }
</script>

<li
  class="lh-copy pv3 ba bl-0 bt-0 br-0 b--dotted b--black-30"
  on:click="{handleClick}"
>
  {title}
</li>
