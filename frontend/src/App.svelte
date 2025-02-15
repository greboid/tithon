<script lang="ts">
  import '../wailsjs/runtime/runtime'
  import './reset.css'
  import './app.css'
  import {NewConnection, UIReady} from '../wailsjs/go/irc/App'
  import { EventsOn, EventsOff, LogInfo } from '../wailsjs/runtime/runtime'
  import {onDestroy, onMount} from 'svelte'
  const servers = $state([])
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
  onMount(() => {
    EventsOn("ServerAdded", (event) => {
      servers.push(event.server)
    })
    EventsOn("ServerUpdated", (event) => {
      const server = servers.find(value => value.id == event.server.id)
      servers.splice(servers.indexOf(server), 1, event.server)
    })
    UIReady()
  })
  onDestroy(() => {
    EventsOff('ServerAdded', 'ServerUpdated')
    servers.length = 0
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
    <h1>Connections</h1>
    <ul>
      {#each servers as server}
        <li>{server.server}</li>
        {:else}
        <li>No Connections</li>
      {/each}
    </ul>
    <button onclick={showModal}>+</button>
  </div>
  <div></div>
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
