<script setup lang="ts">
defineProps<{
  items: Array<{ 
    id?: string
    word: string; 
    translation: string; 
    partOfSpeech?: string; 
    example?: string 
  }>
}>

defineEmits<{
  (e: 'add-favorite', vocabId: string, vocab: { word: string; translation: string }): void
}>()
</script>

<template>
  <div v-if="items.length" class="bg-white rounded-lg shadow p-6">
    <h2 class="text-xl font-semibold mb-4">Extracted Vocabulary</h2>
    <ul class="space-y-3">
      <li
        v-for="(item, i) in items"
        :key="i"
        class="border-b last:border-0 pb-3 last:pb-0 flex items-center justify-between"
      >
        <div class="flex-1">
          <div class="flex justify-between">
            <span class="font-medium">{{ item.word }}</span>
            <span class="text-indigo-600">{{ item.translation }}</span>
          </div>
          <p v-if="item.partOfSpeech" class="text-xs text-gray-500 mt-1">
            {{ item.partOfSpeech }}
          </p>
          <p v-if="item.example" class="text-sm italic text-gray-600 mt-1">
            "{{ item.example }}"
          </p>
        </div>
        <button
          v-if="item.id"
          @click="$emit('add-favorite', item.id, { word: item.word, translation: item.translation })"
          class="text-gray-400 hover:text-red-500 transition text-xl ml-4"
          title="Add to favorites"
        >
          ♡
        </button>
      </li>
    </ul>
  </div>
</template>