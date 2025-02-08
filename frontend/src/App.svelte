<script lang="ts">
  import '../wailsjs/runtime/runtime'
  import './reset.css'
  import './app.css'
  import {EventsOn} from '../wailsjs/runtime/runtime'
  import {GetServers} from '../wailsjs/go/gui/App'
  import Empty from '~lib/Empty.svelte'
  import ServerList from '~lib/ServerList.svelte'
  import ActiveWindow from '~lib/ActiveWindow.svelte'
  import NickList from '~lib/NickList.svelte'

  let servers = $state([])
  EventsOn('serverAdded', server => {
    servers.push(server)
  })
  GetServers().then(response => response.forEach(server => servers.push(server)))
  const {typ, ser, win } = /^\/?(?<typ>[^\/]?)\/?(?<ser>[^\/]*)\/?(?<win>[^\/]*)\/?$/.exec(window.location.pathname)?.['groups'] ?? {typ:"",ser:"",win:""}
</script>
<main>
  {#if servers.length === 0}
    <Empty/>
  {:else}
    <ServerList servers={servers} activeServer={ser} />
    <ActiveWindow activeServer={ser} activeWindow={win}/>
    <NickList/>
  {/if}
</main>
