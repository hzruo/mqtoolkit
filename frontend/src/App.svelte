<script>
  import logo from './assets/images/logo-universal.png'
  import {Greet} from '../wailsjs/go/main/App.js'

  let resultText = "Please enter your name below 👇"
  let name

  function greet() {
    Greet(name).then(result => resultText = result)
  }
</script>

<main class="min-h-screen bg-base-100 flex flex-col items-center justify-center p-4">
  <div class="card w-96 bg-base-200 shadow-xl">
    <div class="card-body items-center text-center">
      <img alt="Wails logo" class="w-32 h-32 mb-4" src="{logo}">
      <h2 class="card-title">MQ Toolkit</h2>
      <p class="text-base-content/70 mb-4">{resultText}</p>
      <div class="form-control w-full max-w-xs">
        <input
          autocomplete="off"
          bind:value={name}
          class="input input-bordered w-full max-w-xs mb-4"
          placeholder="Enter your name"
          type="text"
        />
        <button class="btn btn-primary" on:click={greet}>Greet</button>
      </div>
    </div>
  </div>
</main>


