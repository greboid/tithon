<script>
    import Client from "./Client.svelte";
    import Login from "./Login.svelte"
    import ErrorPanel from "./ErrorPanel.svelte";
    import {loggedIn} from '../stores.js';

</script>
<style>
    :global(html, body) {
        height: 100%;
    }

    :global(body) {
        font: 13px Helmet, Freesans, sans-serif;
        display: flex;
        flex-direction: row;
        gap: 0.5em;
        padding: 0;
        margin: 0;
        align-items: center;
        justify-content: center;
    }
</style>
{#if !$loggedIn}
    <Login/>
{:else}
    <ErrorPanel class="errorPanel"/>
    <Client/>
{/if}