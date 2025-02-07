<script lang="ts">
  import {type Route, Router} from '@mateothegreat/svelte5-router'
  import Home from '~lib/Home.svelte'
  import './reset.css'
  import './app.css'

  const routes: Route[] = [
    {path: '', component: Home},
  ]
</script>

<Router {routes}/>
