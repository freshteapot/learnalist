<script>
  import { link } from "svelte-spa-router";
</script>

<div class="pt2">
  <article
    class="pt5 mw10 center ph3 ph5-ns tc br2 pv5 bg-washed-yellow dark-yellow
    mb5"
  >
    <h1 class="fw3 f3 f2-ns lh-title mt0 mb3">I would like to create a</h1>
    <div class="">
      <a
        class="f6 link dim br1 ba bw1 ph3 pv2 mb2 dib black mr6-ns"
        href="/create/list"
        use:link
      >
        List
      </a>

      <a
        class="f6 link dim br1 ba bw1 ph3 pv2 mb2 dib black"
        href="/create/label"
        use:link
      >
        Label
      </a>
    </div>
  </article>
</div>
