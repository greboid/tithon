<script lang="ts">
  import '../wailsjs/runtime/runtime'
  import './reset.css'
  import './app.css'
  import {EventsOn, EventsOff} from '../wailsjs/runtime/runtime'
  import {Started} from '../wailsjs/go/gui/App'
  import {events} from '../wailsjs/go/models'
  import {onDestroy, onMount} from 'svelte'
  const scrollToBottom = async () => {
    window.scrollTo(0, document.body.scrollHeight);
  };
  const messages = $state([])
  onMount(() => {
    EventsOn('serverAdded', (server: events.ConnectableServer) => {
      messages.push(`Server added: ${server.server}`)
      scrollToBottom();
    })
    EventsOn('channelMessage', (message: events.ChannelMessage) => {
      messages.push(`CM: ${message.target} ${message.source} ${message.message}`)
      scrollToBottom();
    })
    EventsOn('directMessage', (message: events.DirectMessage) => {
      messages.push(`DM: ${message.source} ${message.message}`)
      scrollToBottom();
    })
    EventsOn('channelAdded', (channel: events.Channel) => {
      messages.push(`Channel added: ${channel.name}`)
      scrollToBottom();
    })
  })
  onDestroy(() => {
    EventsOff('serverAdded', 'channelMessage', 'directMessage', 'channelAdded')
  })
  Started()
</script>
<style>
  main {
    display: flex;
    flex-direction: column;
  }
</style>
<main>
  {#each messages as message}
    <p class="message">{message}</p>
  {/each}
</main>
