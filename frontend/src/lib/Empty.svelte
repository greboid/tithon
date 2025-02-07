<script lang="ts">
  import {Connect} from '../../wailsjs/go/gui/App'
  let nickname: string
  let server: string
  let port: number
  let tls: boolean
  let username: string
  let password: string
  const handleSubmit = (e) => {
    e.preventDefault();
    Connect(
        {
          Hostname: server,
          Port: port,
          TLS: tls,
          Saslusername: username,
          Saslpassword: password
        },
        {
          Nick: nickname
        }
    )
  }
</script>

<style>
  div {
    width: 100%;
    height: 100%;
    display: flex;
    justify-content: center;
    align-items: center;
  }
</style>
<div>
  <form on:submit={handleSubmit}>
    <label for="nickname">Nickname</label>
    <input type="text" name="nickname" bind:value={nickname} />
    <label for="server">Server</label>
    <input type="text" name="server" bind:value={server} />
    <label for="port">Port</label>
    <input type="number" name="port" bind:value={port} />
    <label for="tls">TLS</label>
    <input type="checkbox" name="tls" bind:checked={tls} />
    <label for="saslusername">Username</label>
    <input type="text" name="saslusername" bind:value={username} />
    <label for="saslpassword">Password</label>
    <input type="text" name="saslpassword" bind:value={password} />
    <button>Connect</button>
  </form>
</div>
