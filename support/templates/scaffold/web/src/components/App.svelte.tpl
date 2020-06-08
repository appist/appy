<script lang="ts">
  import router from "page";
  import { isLoading } from "svelte-i18n";

  let route, params;

  router("/", () =>
    import(/* webpackChunkName: 'home-page' */ "@/pages/Home.svelte")
      .then(res => (route = res.default))
      .catch(err => {})
  );

  router.start();
</script>

<style>
  :global(html, body, #app) {
    height: 100%;
  }
</style>

<template>
  {#if !$isLoading}
    <svelte:component this={route} {params} />
  {/if}
</template>
