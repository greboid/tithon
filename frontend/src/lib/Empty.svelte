<script lang="ts">
  import {Connect} from '../../wailsjs/go/gui/App'
  let nickname: string = $state("")
  let server: string = $state("")
  let port: number = $state(6667)
  let tls: boolean = $state(true)
  let username: string = $state("")
  let password: string = $state("")

  let connecting = $state(false)
  let errorMessage = $state("")

  const handleSubmit = (e: SubmitEvent): void => {
    e.preventDefault();
    connecting = true
    errorMessage = ""
    Connect(
        {
          server: `${server}:${port}`,
          tls: tls,
          saslUsername: username,
          saslPassword: password,
          profile: {
            nick: nickname,
          },
        }
    ).catch(error => {
      connecting = false
      errorMessage = error
    })
  }
</script>

<div>
  <form onsubmit={handleSubmit}>
    <label for="nickname">Nickname</label>
    <input type="text" name="nickname" bind:value={nickname} required />
    <label for="server">Server</label>
    <input type="text" name="server" bind:value={server} required />
    <label for="port">Port</label>
    <input type="number" name="port" bind:value={port} />
    <label for="tls">TLS</label>
    <input type="checkbox" name="tls" bind:checked={tls} required/>
    <label for="saslusername">Username</label>
    <input type="text" name="saslusername" bind:value={username} />
    <label for="saslpassword">Password</label>
    <input type="password" name="saslpassword" bind:value={password} />
    <button disabled={connecting}>Connect</button>
  </form>
  <p>{errorMessage}</p>
</div>

<style>
  div {
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
  }
  p {
    padding-top: 1em;
  }
</style>
