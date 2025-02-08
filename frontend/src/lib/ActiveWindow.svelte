<script>
  let { activeServer, activeWindow } = $props()
</script>
{#if activeServer}
  <p>A server is active</p>
{/if}
{#if activeWindow}
  <p>A Window is active</p>
{/if}
