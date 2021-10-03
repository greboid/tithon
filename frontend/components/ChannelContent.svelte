<script>
    import { messages } from "../stores";
</script>
<div class="channel">
    {#each $messages as message}
        <p>{message.Time} | &lt;{message.Source.Name}&gt; {message.Message}</p>
    {/each}
</div>