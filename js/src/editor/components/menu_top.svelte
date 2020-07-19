<script>
  import { link } from "svelte-spa-router";
</script>

<style>
  @import "../../../all.css";
</style>

<header class="flex">
  <div class="w-25 pa3 mr2" />

  <div class="w-75 pa3 items-end">
    <div class="fr mt0">
      <a
        title="Documentation"
        href="/create"
        use:link
        class="f6 fw6 hover-blue link black-70 mr2 mr3-m mr2-l di">
        Create
      </a>
      <a
        title="Components"
        href="/lists/by/me"
        use:link
        class="f6 fw6 hover-blue link black-70 mr2 mr3-m mr5-l di">
        Find
      </a>
    </div>
  </div>
</header>
