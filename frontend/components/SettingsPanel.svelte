<script>
    import MdAdd from 'svelte-icons/md/MdAdd.svelte'
    import MdEdit from 'svelte-icons/md/MdEdit.svelte'
    import MdSettings from 'svelte-icons/md/MdSettings.svelte'
    import {addNetwork} from "../stores";

    const showAddNetwork = () => {
        addNetwork.set(true)
    }
</script>
<style>
    button {
        width: 2em;
        height: 2em;
    }
</style>
<button title="Edit Networks" on:click={showAddNetwork}><MdAdd/></button>
<button title="Edit Profiles"><MdEdit/></button>
<button title="Edit Settings"><MdSettings /></button>