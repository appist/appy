<script>
  import { createEventDispatcher } from 'svelte'

  export let text = ''

  const dispatch = createEventDispatcher()

  function onClick(event) {
    dispatch('click', event)
  }
</script>

<style>
  .button {
    border: 1px solid #eee;
    border-radius: 3px;
    background-color: #ffffff;
    cursor: pointer;
    font-size: 15px;
    padding: 3px 10px;
    margin: 10px;
  }
</style>

<button class="button" on:click={onClick}>{text}</button>
