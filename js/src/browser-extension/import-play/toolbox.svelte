<script>
  import { push } from "svelte-spa-router";
  const toolbox = [
    {
      href: "/toolbox/read-write-repeat",
      description: `Read, write, repeat`
    },
    {
      href: "/start",
      description: `Import and play`
    },
    {
      href: "/settings",
      description: `Settings`
    }
  ];
</script>

<div class="flex flex-column">
  <div class="w-100 pa3 mr2">
    <ul>
      {#each toolbox as tool}
        <li on:click|preventDefault={() => push(tool.href)}>
          <a href="#" class="black-80">{tool.description}</a>
        </li>
      {/each}
    </ul>
  </div>
</div>
