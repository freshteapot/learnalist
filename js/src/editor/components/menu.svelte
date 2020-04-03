<script>
  import { push, link } from "svelte-spa-router";
  let isLoggedIn = true;
</script>

<header class="pt2 dt w-100">
  <div class="dtc v-mid tc white ph3 ph4-l">
    <button on:click="{() => push('/lists/by/me')}">My Lists</button>
    <button on:click="{() => push('/labels/by/me')}">My Labels</button>
    <span>/</span>
    <button on:click="{() => push('/list/new')}">New List</button>
    <button on:click="{() => push('/list/new')}">New Label</button>
  </div>
</header>
