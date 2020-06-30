<script lang="ts">
  import { _ } from "svelte-i18n";
</script>

<style>
  * {
    text-align: center;
  }

  h1 {
    margin-top: 20rem;
  }
</style>

<template>
  <h1>{$_('welcome.heading')}</h1>

  <p>An opinionated productive web framework that helps scaling business easier.</p>
</template>
