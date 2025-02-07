<script lang="ts">
  import ServerList from '~lib/ServerList.svelte'
  import ActiveWindow from '~lib/ActiveWindow.svelte'
  import NickList from '~lib/NickList.svelte'
  import Empty from '~lib/Empty.svelte'
  let { servers  = []} = $props()
</script>

{#if servers.length === 0}
   <Empty />
{:else}
  <ServerList servers={servers} />
  <ActiveWindow />
  <NickList />
{/if}
