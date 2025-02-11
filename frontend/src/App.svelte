<script lang="ts">
  import '../wailsjs/runtime/runtime'
  import './reset.css'
  import './app.css'
  import {EventsOn, LogInfo} from '../wailsjs/runtime/runtime'
  import {Started} from '../wailsjs/go/gui/App'
  import * as irc from '../wailsjs/go/models'
  const scrollToBottom = async () => {
    window.scrollTo(0, document.body.scrollHeight);
  };
  const messages = $state([])
  EventsOn('serverAdded', (server: irc.irc.ConnectableServer) => {
    messages.push(`Server added: ${server.server}`)
    scrollToBottom();
  })
  EventsOn('channelMessage', (message: irc.irc.ChannelMessage) => {
    messages.push(`CM: ${message.target} ${message.source} ${message.message}`)
    scrollToBottom();
  })
  EventsOn('directMessage', (message: irc.irc.DirectMessage) => {
    messages.push(`DM: ${message.source} ${message.message}`)
    scrollToBottom();
  })
  EventsOn('channelAdded', (channel: irc.irc.Channel) => {
    messages.push(`Channel added: ${channel.name}`)
    scrollToBottom();
  })
  const {typ, ser, win } = /^\/?(?<typ>[^\/]?)\/?(?<ser>[^\/]*)\/?(?<win>[^\/]*)\/?$/.exec(window.location.pathname)?.['groups'] ?? {typ:"",ser:"",win:""}
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
    <p>{message}</p>
  {/each}
</main>
