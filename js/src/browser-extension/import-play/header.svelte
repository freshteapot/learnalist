<script>
  import { push } from "svelte-spa-router";
</script>

<header class="flex pv2 w-100 justify-between bb b--black-10">

  <div class="w-25 pa1">
    <a
      href="#"
      class="f5 f4-ns fw6 link black-20"
      title="A Learnalist creation">
      <small class="black-40">Learnalist</small>
    </a>
  </div>
  <div class="w-25 pa1">
    <a
      href="#"
      on:click|preventDefault={() => push('/toolbox')}
      class="f5 f4-ns fw6 link black-20 "
      title="Toolbox">
      <small class="black-70">Toolbox</small>
    </a>

    <a
      href="#"
      on:click|preventDefault={() => push('/settings')}
      class="f5 f4-ns fw6 link black-20"
      title="Settings">
      <small class="black-70">Settings</small>
    </a>
  </div>

</header>
