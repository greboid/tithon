<script lang="ts">
  import {EventsOn, LogInfo} from '../../wailsjs/runtime'
  import {ircmsg} from '../../wailsjs/go/models'
  let text = ""
  EventsOn("privmsg", (message: ircmsg.Message) => {
    LogInfo(message.Source + ": " + message.Params.join(" "))
  })
</script>

<p>{text}</p>
