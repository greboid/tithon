<script>
    import {connected, addNetwork} from '../stores.js';

    export let doAddNetwork

    let name, server, tls = true, nickname, realname, username, password

    const dismiss = () => {
        addNetwork.set(false)
        name = ""
        server = ""
        tls = true
        nickname = ""
        realname = ""
        username = ""
        password = ""
    }
    const escapeHandler = event => {
        if (event.key === 'Escape') {
            dismiss()
        }
    }
    const handleAdd = () => {
        doAddNetwork(name, server, tls, nickname, realname, username, password)
        dismiss()
    }
</script>
<style>
    :global(#addNetwork) {
        position: absolute;
        width: 99%;
        height: 99%;
        backdrop-filter: blur(1px);
        display: flex;
        justify-content: center;
        align-items: center;
    }

    .content {
        border: 1px solid black;
        padding: 1em;
    }

    p {
        padding: 1em;
        text-align: center;
    }

    form {
        display: grid;
        grid-template-columns: [labels] auto [controls] 1fr;
        grid-auto-flow: row;
        grid-gap: .8em;
        padding: 1.2em;
    }

    form input {
        justify-self: flex-start;
    }

    form input[type="checkbox"] {
        margin: 0;
    }

    form > section {
        display: flex;
    }

    form > section > button {
        flex-grow: 2;
    }
</style>
<svelte:window on:keydown={escapeHandler} />
{#if $connected && $addNetwork}
    <section id="addNetwork">
        <div class="content">
            <p>Add Network</p>
            <form on:submit|preventDefault={handleAdd}>
                <label>Name:</label><input type="text" bind:value="{name}" required />
                <label>Server:</label><input type="text" bind:value="{server}" required />
                <label>TLS</label><input type="checkbox" bind:checked="{tls}" />
                <label>Nickname:</label><input type="text" bind:value="{nickname}" required />
                <label>Realname:</label><input type="text" bind:value="{realname}" required />
                <label>Username:</label><input type="text" bind:value="{username}" />
                <label>Password:</label><input type="password" bind:value="{password}" />
                <div></div>
                <section>
                    <button on:click={dismiss}>Cancel</button>
                    <button type="submit">Add</button>
                </section>
            </form>
        </div>
    </section>
{/if}
