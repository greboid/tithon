<script lang="ts">
  import sanitizeHtml from 'sanitize-html';
  import '../wailsjs/runtime/runtime'
  import './reset.css'
  import './app.css'
  import { events } from '../wailsjs/go/models'
  import {NewConnection, UIReady} from '../wailsjs/go/irc/App'
  import { EventsOn, EventsOff, LogInfo, BrowserOpenURL } from '../wailsjs/runtime'
  import {onDestroy, onMount} from 'svelte'

  document.body.addEventListener("click", function (e) {
    if (e.target && e.target.nodeName === "A" && !e.target.href.startsWith('wails')) {
      e.preventDefault();
      console.log("Capturing link:", e.target.innerText);
      BrowserOpenURL(e.target.href);
    }
  });
  const servers: events.Server[] = $state([])
  let activeServer: events.Server = $state(null)
  let activeChannel: events.Channel = $state(null)
  let messages: events.ChannelMessage[] = $state([])
  const showModal = () => {
    document.getElementsByTagName('dialog')[0].returnValue = ""
    document.getElementsByTagName('dialog')[0].showModal()
  }
  const hideDialog = () => {
    document.getElementsByTagName('dialog')[0].close("false")
  }
  const handleAddNewConnection = (e) => {
    e.preventDefault()
    if (document.getElementsByTagName('dialog')[0].returnValue === "false") {
      return
    }
    const data = Object.fromEntries(new FormData(e.target).entries())
    NewConnection(data.url, data.usetls==="on", data.username, data.password, data.nickname)
    .then(() => document.getElementsByTagName('dialog')[0].close("false"))
    .catch(error => alert(error))
  }
  const handleServerSelect = (event, server: events.Server) => {
    event.preventDefault()
    activeChannel = null
    activeServer = server
    messages.length = 0
  }
  const handleChannelSelect = (e, server, channel) => {
    e.preventDefault()
    activeChannel = channel
    messages.length = 0
  }
  const urlify = (text: string): string => {
    text = sanitizeHtml(text, {
      allowedTags: [],
      disallowedTagsMode: 'recursiveEscape',
    })
    const results = /https?:\/\/\S*/ig.exec(text)
    results.forEach(result => {
      text = text.replaceAll(result, `<a href="${result}">${result}</a>`)
    })
    return text
  }
  onMount(() => {
    EventsOn("ChannelMessageReceived", (event: events.ChannelMessageReceived) => {
      if (activeChannel != null && activeChannel.name == event.message.channel.name) {
        messages.push(event.message)
      }
    })
    EventsOn("ServerAdded", (event: events.ServerAdded) => {
      servers.push(event.server)
      if (servers.length === 1) {
        activeServer = event.server
      }
    })
    EventsOn("ServerUpdated", (event: events.ServerUpdated) => {
      const server = servers.find(value => value.id == event.server.id)
      servers.splice(servers.indexOf(server), 1, event.server)
    })
    EventsOn("ChannelJoinedSelf", (event: events.ChannelJoinedSelf) => {
      const server = servers.find(value => value.id == event.channel.serverid)
      server.channels.push(event.channel)
      server.channels.sort((a, b) => a.name < b.name ? -1 : 1)
    })
    UIReady()
  })
  onDestroy(() => {
    EventsOff('ServerAdded', 'ServerUpdated', 'ChannelJoinedSelf')
    servers.length = 0
    activeServer = null
    activeChannel = null
  })
</script>
<style>
  h2 {
    margin: 0;
    padding: 0 0 1em;
  }
  main {
    display: flex;
    flex-direction: row;
    gap: 1em;
  }
  .sl {
    display: flex;
    flex-direction: column;
    padding-right: 1em;
    border-right: 1px solid grey;
    ul {
      padding: 0;
      margin: 0;
      & li {
        list-style: none;
        padding: 0;
        margin: 0;
      }
      & ul {
        padding-left: 1em;
      }
      }
  }
  main:has(dialog[open]) {
    filter: blur(4px);
  }
  form {
    grid-template-columns: 5em 1fr;
  }
</style>
<main>
  <div class="sl">
    <ul>
      {#each servers as server}
        <li>
          <a onclick={(e) => handleServerSelect(e, server)} href="/{server.id}">{server.server}</a>
          <ul>
            {#each server.channels as channel}
              <li><a onclick={(e) => handleChannelSelect(e, server, channel)} href="/{server.id}/{channel.name}">{channel.name}</a></li>
            {/each}
          </ul>
        </li>
        {:else}
        <li>No Connections</li>
      {/each}
    </ul>
    <button onclick={showModal}>+</button>
  </div>
  <div>
    {#if activeChannel === null}
      {#if activeServer === null}
        <p>No Server</p>
      {:else}
        <p>{activeServer.id}</p>
      {/if}
    {:else}
      <h1>{activeChannel.name}</h1>
      {#each messages as message}
        <p>{message.source.nick}: {@html urlify(message.message)}</p>
        {/each}
    {/if}
  </div>
  <dialog>
    <h2>Add new Connection</h2>
    <form onsubmit={handleAddNewConnection}>
      <label for="usetls">Use TLS</label>
      <input type="checkbox" name="usetls" checked={true} />
      <label for="url">URL</label>
      <input type="text" name="url"/>
      <label for="nickname">Nickname</label>
      <input type="text" name="nickname"/>
      <label for="username">Username</label>
      <input type="text" name="username"/>
      <label for="password">Password</label>
      <input type="password" name="password"/>
      <button type="submit">Add</button>
      <button onclick={hideDialog}>Cancel</button>
    </form>
  </dialog>
</main>
