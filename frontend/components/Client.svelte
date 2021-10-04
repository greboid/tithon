<script>
    import ServerList from "./ServerList.svelte";
    import NickList from "./NickList.svelte";
    import SubjectPanel from "./SubjectPanel.svelte";
    import ChannelContent from "./ChannelContent.svelte";
    import WindowInput from "./WindowInput.svelte";
    import ReconnectingWebSocket from "reconnecting-websocket";
    import {connected, serverList, messages} from "../stores";
    import SettingsPanel from "./SettingsPanel.svelte";
    import NetworkEditor from "./NetworkEditor.svelte";

    let socket
    const startSocket = socketLocation => {
        socket = new ReconnectingWebSocket(socketLocation)
        socket.onopen = () => {
            connected.update(() => true)
            socket.send(JSON.stringify({
                "action": "INIT",
                "message": {
                    "since": 0,
                },
            }))
        }
        socket.onmessage = (event) => {
            let eventJSON = JSON.parse(event.data)
            if (eventJSON.hasOwnProperty("serverlist")) {
                let map = new Map()
                Object.keys(eventJSON["serverlist"]).forEach(function(key) {
                    eventJSON["serverlist"][key].sort()
                    map.set(key, eventJSON["serverlist"][key])
                })
                serverList.update(() => map)
            }
            if (eventJSON.hasOwnProperty("message")) {
                messages.update(old => {
                    let date = new Date(eventJSON["message"].Time * 1000)
                    eventJSON["message"].Time = date.toLocaleTimeString('en-gb', {hour: '2-digit', minute: '2-digit', second: '2-digit' })
                    old.push(eventJSON["message"])
                    return old
                })
            }
        }
        socket.onclose = () => {
            connected.update(() => false)
        }
    }
    if (location.protocol === "https") {
        startSocket("wss://" + window.location.host + "/socket")
    } else {
        startSocket("ws://" + window.location.host + "/socket")
    }
    const sendToIRC = (network, channel, message) => {
        socket.send(JSON.stringify({
            "action": "SENDCHANMESSAGE",
            "message": {
                "network": network,
                "channel": channel,
                "message": message,
            },
        }))
    }
    const doAddNetwork = (name, server, tls, nickname, realname, username, password) => {
        socket.send(JSON.stringify({
            "action": "ADDNETWORK",
            "message": {
                "name": name,
                "server": server,
                "tls": tls,
                "nickname": nickname,
                "realname": realname,
                "username": username,
                "password": password,
            },
        }))
    }
</script>
<style>
    main {
        width: 100vw;
        height: 100vh;
        padding: 0.5em;
        display: grid;
        grid-template-columns: 15em auto;
        grid-template-rows: 1fr;
        grid-auto-columns: 1fr;
        gap: 0 0;
        grid-auto-flow: row;
        grid-template-areas:
            "nav main"
            "settings main"
    }

    .serverlist {
        grid-area: nav;
    }

    .main {
        display: grid;
        grid-template-columns: 1fr 0.2fr;
        grid-template-rows: 2em 1fr 2em;
        gap: 0.5em 0;
        grid-auto-flow: row;
        grid-template-areas:
            "subject nick"
            "channel nick"
            "input nick";
        grid-area: main;
    }

    .settings {
        grid-area: settings;
    }

    .nick {
        grid-area: nick;
    }

    .channel {
        grid-area: channel;
        overflow: auto;
    }

    .input {
        grid-area: input;
    }
</style>
<main>
    <NetworkEditor doAddNetwork="{doAddNetwork}"/>
    <section class="serverlist">
        <ServerList />
    </section>
    <section class="settings">
        <SettingsPanel />
    </section>
    <div class="main">
        <section class="nick">
            <NickList />
        </section>
        <section class="subject">
            <SubjectPanel/>
        </section>
        <section class="channel">
            <ChannelContent />
        </section>
        <section class="input">
            <WindowInput {sendToIRC} />
        </section>
    </div>
</main>