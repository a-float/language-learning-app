<script setup lang="ts">
import { ref } from 'vue'

const responses = ref<string[]>([])

async function handleSubmit(e: SubmitEvent) {
  const form = e.target as HTMLFormElement
  const formData = new FormData(form)
  const text = formData.get('text') as string
  const res = await fetch('http://localhost:8090/api/answer', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ text, targetLanguage: 'Ënglish' }),
  })

  if (!res.ok) return
  const { answer } = await res.json()
  console.log(res)
  responses.value.push(answer)
}
</script>

<template>
  <h1>Create learning material</h1>
  <form @submit.prevent="handleSubmit">
    <label>
      Message to ask my ai
      <textarea defaultValue="Hello AI!" name="text"></textarea>
    </label>
    <button>Submit</button>
  </form>
  <ul>
    <li v-for="item in responses" style="background-color: #ccddff; margin-bottom: 2rem">
      <pre>{{ JSON.stringify(JSON.parse(item), null, 4) }}</pre>
    </li>
  </ul>
</template>

<style scoped></style>
