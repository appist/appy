<script>
let count = 0

function handleClick(num) {
  count += num
}
</script>

<style lang="scss" global>
@import 'bootstrap/scss/bootstrap';
</style>

<template lang="pug">
  div Hello World!
  button(class="btn-primary" on:click!="{() => handleClick(-1)}") -
  span {count}
  button(class="btn-primary" on:click!="{() => handleClick(1)}") +
</template>
