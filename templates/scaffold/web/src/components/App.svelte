<script lang="ts">
  import router from "page";

  let route, params;

  router("/", () =>
    import(/* webpackChunkName: 'home-page' */ "@/pages/Home.svelte")
      .then(res => (route = res.default))
      .catch(err => {})
  );

  router("*", () => {
    import(/* webpackChunkName: 'not-found-page' */ "@/pages/NotFound.svelte")
      .then(res => (route = res.default))
      .catch(err => {});
  });

  router.start();
</script>

<style lang="scss" global>
  @import "bootstrap/scss/functions";
  @import "bootstrap/scss/mixins";
  @import "@assets/styles/theme/variables";
  @import "bootstrap/scss/bootstrap";

  html,
  body,
  #app {
    height: 100%;
  }
</style>

<template lang="pug" src="./App.pug" />
