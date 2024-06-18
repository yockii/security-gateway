<script lang="ts" setup>
import {Port} from '@/types/service'

defineProps<{
  selectedPort: number | undefined
  ports: Port[]
}>()
const emit = defineEmits(['portSelected'])


</script>

<template>
  <a-list hoverable split>
    <template #header>
      <div>端口列表</div>
    </template>
    <a-list-item v-for="port in ports" @click="emit('portSelected', port.port)">
      <div :class="{ 'font-600 font-italic': selectedPort === port.port }"
           class="flex items-center justify-between cursor-pointer">
        <span>{{ port.port }}</span>
        <div v-if="port.inUse" class="w-10px h-10px b-rd-50% bg-green"></div>
        <div v-else class="w-10px h-10px b-rd-50% bg-gray"></div>
      </div>
    </a-list-item>
  </a-list>
</template>