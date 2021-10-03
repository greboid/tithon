<script>
    import {connected} from '../stores.js';
</script>
<style>
    :global(#errorPanel) {
        position: absolute;
        width: 100%;
        height: 100%;
        backdrop-filter: blur(1px);
        display: flex;
        justify-content: center;
        align-items: center;
    }
    .content {
        border: 1px solid red;
    }
    p {
        padding: 1em;
        text-align: center;
    }
</style>
{#if !$connected}
    <section id="errorPanel">
        <div class="content">
            <p>No backend connection</p>
        </div>
    </section>
{/if}
