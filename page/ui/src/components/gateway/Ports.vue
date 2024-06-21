<script lang="ts" setup>
import {getAllPorts, getServiceList} from '@/api/service';
import {Port, Service} from '@/types/service';
import {computed, onMounted, onUnmounted, reactive, ref} from 'vue';
import Services from './Services.vue';

const ports = ref<Port[]>([])
const sortedPorts = computed(() => {
  return ports.value.sort((a, b) => a.port - b.port)
})
const selectedPort = ref<Port | undefined>(undefined)
const getPorts = async () => {
  try {
    const resp = await getAllPorts()
    ports.value = resp.data || []
  } catch (error) {
    console.log(error)
  }
}
const portSelected = (port: Port) => {
  if (port && port.port === selectedPort.value?.port) {
    return
  }
  selectedPort.value = port
  getServiceInPort()
}

const serviceList = ref<Service[]>([])
const servicePaginationProps = reactive({
  defaultPageSize: 10,
  total: 0
})
const servicePage = ref<number>(1)
const getServiceInPort = async () => {
  if (!selectedPort.value) {
    return
  }
  try {
    const resp = await getServiceList({port: selectedPort.value.port, page: servicePage.value})
    serviceList.value = resp.data?.items || []
    servicePaginationProps.total = resp.data?.total || 0
  } catch (error) {
    console.log(error)
  }
}
const servicePageChanged = (page?: number) => {
  if (page) {
    servicePage.value = page
  }
  getServiceInPort()
}

// 定时更新端口列表
const timer = ref<number | undefined>(undefined)

onMounted(() => {
  timer.value = window.setInterval(() => {
    getPorts()
  }, 5000)
  getPorts()
})

onUnmounted(() => {
  if (timer.value) {
    window.clearInterval(timer.value)
  }
})
</script>

<template>
  <a-split class="h-100% w-100%" default-size="160px" max="200px" min="120px">
    <template #first>
      <a-list class="mr-8px" hoverable split>
        <template #header>
          <div>端口列表</div>
        </template>
        <a-list-item v-for="port in sortedPorts" :key="port.port" @click="portSelected(port)">
          <div class="flex items-center justify-between cursor-pointer">
            <span :class="{ 'font-italic text-blue font-600': selectedPort && selectedPort.port === port.port }">{{
                port.port
              }}</span>
            <div v-if="port.inUse" class="w-10px h-10px b-rd-50% bg-green"></div>
            <div v-else class="w-10px h-10px b-rd-50% bg-gray"></div>
          </div>
        </a-list-item>
      </a-list>
    </template>
    <template #second>
      <Services :pagination-props="servicePaginationProps" :service-list="serviceList"
                @page-changed="servicePageChanged" @port-changed="getPorts"/>
    </template>
  </a-split>
</template>