<script lang="ts">
  let { servers } = $props()
</script>
<nav>
  <h2>Servers</h2>
  <ul>
    {#each servers as server}
      <li>{server.name}</li>
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
