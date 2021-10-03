<script>
    import { loggedIn } from '../stores.js';
    let error = false
    function login(e) {
        const formData = new FormData(e.target);
        if (formData.get("username") === "test" && formData.get("password") === "test") {
            loggedIn.update(_ => true)
            error = false
        } else {
            error = true
        }
    }
</script>
<style>
    main {
        display: flex;
        justify-content: center;
        flex-direction: column;
    }
    form {
        display: grid;
        grid-template-columns: [labels] auto [controls] 1fr;
        grid-auto-flow: row;
        grid-gap: .8em;
        padding: 1.2em;
    }
    label {
        grid-column: labels;
        grid-row: auto;
    }
    input, button {
        grid-column: controls;
        grid-row: auto;
    }
</style>
<main>
    <p>Currently the login are hardcoded and have no real effect.</p>
    <form on:submit|preventDefault="{login}">
        {#if error}
            <p>Incorrect details</p>
        {/if}
        <label for="username">Username:</label>
        <input type="text" id="username" name="username" required value="test">
        <br>
        <label for="password">Password:</label>
        <input type="password" id="password" name="password" required value="test">
        <button type="submit">Login</button>
    </form>
</main>