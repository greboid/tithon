<script lang="ts">
  let { servers, activeServer } = $props()
</script>
<nav>
  <h2>Servers</h2>
  <ul>
    {#each servers as server}
      <li><a href="/s/{server.name}">{server.name}</a></li>
    {:else}
      <li>No servers</li>
    {/each}
  </ul>
</nav>
<style>
  ul {
    list-style: none;
    padding-left: 0;
  }
</style>
