<script>
  import version from "../../stores/version";
  const { loading, error } = version;
  version.get();
</script>

<div class="pa3 pa5-ns">
  <div class="pl0 measure center">
    {#if $error}
      error is {$error}
    {:else if $loading}
      Loading...
    {:else}
      <p>Version is {$version.version}</p>
      <p>Git hash is {$version.gitHash}</p>
      <p>Git date is {$version.gitDate}</p>
      <p>
        On
        <a href={$version.url} target="_blank">Github</a>
      </p>
    {/if}
  </div>
</div>
