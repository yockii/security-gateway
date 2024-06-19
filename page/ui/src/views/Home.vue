<script lang="ts" setup>
import { getRouteWithTargetList } from '@/api/route';
import { getAllPorts, getServiceList } from '@/api/service';
import { RouteWithTarget } from '@/types/route';
import { Port, Service } from '@/types/service';
import { onMounted, ref } from 'vue';

import PortList from '@/components/PortList.vue';
import ServiceList from '@/components/ServiceList.vue';
import RouteList from '@/components/RouteList.vue';

const ports = ref<Port[]>([])
const getPorts = async () => {
  try {
    const resp = await getAllPorts()
    ports.value = resp.data || []
  } catch (error) {
    console.log(error)
  }
}

const selectedPort = ref<number | undefined>(undefined)

const portSelected = (port: number) => {
  selectedPort.value = port
  serviceList.value = []
  selectedService.value = undefined
  routes.value = []
  selectedRoute.value = undefined
  getServiceInPort()
}


const serviceList = ref<Service[]>([])
const selectedService = ref<Service | undefined>(undefined)
const serviceTotal = ref<number>(0)
const getServiceInPort = async () => {
  if (!selectedPort.value) {
    return
  }
  try {
    const resp = await getServiceList({ port: selectedPort.value })
    serviceList.value = resp.data?.items || []
    serviceTotal.value = resp.data?.total || 0
  } catch (error) {
    console.log(error)
  }
}
const serviceSelected = (service: Service) => {
  selectedService.value = service
  getRouteList()
}
const serviceUpdated = (service: Service | undefined) => {
  // 判断端口是否有新增
  const updatedPort = !service || (!ports.value.find((port) => port.port === service.port))
  if (updatedPort) {
    getPorts()
  }
  // 重新获取端口下的服务列表
  if (selectedPort.value) {
    getServiceInPort()
  }
}


const selectedRoute = ref<string | undefined>(undefined)
const routeSelected = (route: RouteWithTarget) => {
  selectedRoute.value = route.uri
}

const routes = ref<RouteWithTarget[]>([])

const getRouteList = async () => {
  if (!selectedService.value) return
  try {
    const resp = await getRouteWithTargetList({ serviceId: selectedService.value?.id })
    routes.value = resp.data.items
  } catch (error) {
    console.log(error)
  }
}
const routeUpdated = () => {
  getRouteList()
}

onMounted(() => {
  getPorts()
});
</script>

<template>
  <a-layout class="p-16px">
    <!-- 端口列表 -->
    <a-layout-sider style="width: 120px;">
      <PortList :ports="ports" :selectedPort="selectedPort" @port-selected="portSelected" />
    </a-layout-sider>
    <!-- 服务列表 -->
    <a-layout-sider style="min-width: 240px; margin-left: 1px;">
      <ServiceList :selected-service="selectedService" :service-list="serviceList" @service-selected="serviceSelected"
        @service-updated="serviceUpdated" />
    </a-layout-sider>
    <!-- 路由列表 -->
    <a-layout-content>
      <RouteList :routes="routes" :selected-route="selectedRoute" :selected-service="selectedService"
        @route-selected="routeSelected" @route-updated="routeUpdated" />
    </a-layout-content>
  </a-layout>

</template>