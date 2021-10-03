<script>
    import {selectedChannel, selectedNetwork, serverList} from "../stores";

const selectWindow = event => {
    selectedNetwork.update(_ => event.target.dataset.network)
    selectedChannel.update(_ => event.target.dataset.channel)
}
</script>
<style>
    ul.networks {
        list-style: none;
        padding: 0;
        margin: 0;
    }
    ul.windows {
        list-style: none;
        padding: 0 0 0 1em;
        margin: 0;
    }
    li {
        padding: 0.2em;
    }
    a.active {
        font-weight: bold;
    }
</style>
<nav>
    {#if [...$serverList.keys()].length > 0}
    <ul class="networks">
        {#each [...$serverList.keys()] as network}
            <li>
                <a data-network="{network}" data-channel="" on:click|preventDefault={selectWindow}
                   class:active="{$selectedNetwork === network && $selectedChannel === ''}"
                >{network}</a>
                {#if [...$serverList.get(network).values()].length > 0}
                    <ul class="windows">
                        {#each [...$serverList.get(network).values()] as window}
                            <li><a data-network="{network}" data-channel="{window}" on:click|preventDefault={selectWindow}
                                   class:active="{$selectedNetwork === network && $selectedChannel === window}"
                            >{window}</a></li>
                        {/each}
                    </ul>
                {/if}
            </li>
        {/each}
    </ul>
    {:else}
        <p>No networks</p>
    {/if}
</nav>