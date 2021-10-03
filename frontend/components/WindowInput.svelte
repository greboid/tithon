<script>
    import {selectedChannel, selectedNetwork} from "../stores";

    export let sendToIRC
    let newMessage = ""
    const sendMessage = () => {
        sendToIRC($selectedNetwork, $selectedChannel, newMessage)
        newMessage = ""
    }
</script>
<style>
    input {
        width: 100%;
    }
</style>
{#if $selectedChannel !== ""}
<form on:submit|preventDefault={sendMessage}>
    <input class="input" type="text" bind:value={newMessage}>
</form>
{/if}