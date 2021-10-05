<script>
    import {selectedChannel, selectedNetwork} from "../stores";

    export let sendToIRC, joinChannel
    let newMessage = ""
    const parseCommand= () => {
        if (newMessage.startsWith("/join ")) {
            joinChannel($selectedNetwork, newMessage.substring(6))
        }
        newMessage = ""
    }
    const sendMessage = () => {
        if (newMessage.startsWith("/")) {
            parseCommand()
            return
        }
        if ($selectedChannel !== "") {
            sendToIRC($selectedNetwork, $selectedChannel, newMessage)
            newMessage = ""
        }
    }
</script>
<style>
    input {
        width: 100%;
    }
</style>
<form on:submit|preventDefault={sendMessage}>
    <input class="input" type="text" bind:value={newMessage}>
</form>