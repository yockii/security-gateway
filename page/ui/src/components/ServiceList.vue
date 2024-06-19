<script lang="ts" setup>
import {addService, deleteService, updateService} from '@/api/service';
import {Service} from '@/types/service';
import {Message} from '@arco-design/web-vue';
import {ref} from 'vue';

import UserInfoRouteModal from './UserInfoRouteModal.vue';
import ServiceFieldDrawer from './ServiceFieldDrawer.vue';

defineProps<{
  selectedService: Service | undefined
  serviceList: Service[]
}>()

const emit = defineEmits(['serviceSelected', 'serviceUpdated'])

const currentService = ref<Service>({})
const showServiceModal = ref(false)

const editService = (service: Service) => {
  // 复制一份，避免直接修改
  currentService.value = JSON.parse(JSON.stringify(service))
  showServiceModal.value = true
}
const handleServiceEditor = async (done: (closed: boolean) => void) => {
  let resp = null
  if (!currentService.value.id) {
    // 新增
    // 检查服务名称是否为空
    if (!currentService.value.name || !currentService.value.port) {
      Message.warning('服务名称及端口不能为空')
      return
    }
    try {
      resp = await addService(currentService.value)
    } catch (error) {
      console.log(error)
      Message.error('新增服务失败')
      return
    }
  } else {
    try {
      resp = await updateService(currentService.value)
    } catch (error) {
      console.log(error)
      Message.error('更新服务失败')
      return
    }
  }
  if (resp.code === 0) {
    Message.success('操作成功')
    done(true)
    emit('serviceUpdated', resp.data)
  } else {
    Message.error('操作失败')
  }
}

// 删除服务
const deleteServ = async (service: Service) => {
  try {
    const resp = await deleteService(service.id!)
    if (resp.code === 0) {
      Message.success('删除成功')
      emit('serviceUpdated')
    } else {
      Message.error('删除失败')
    }
  } catch (error) {
    console.log(error)
    Message.error('删除失败')
  }
}

const showDesensitiveDrawer = ref(false)

// 拦截用户信息配置
const userInfoRouteService = ref<Service>()
const showUserInfoRouteModal = ref(false)
const editServiceUserRoute = (service: Service) => {
  userInfoRouteService.value = service
  showUserInfoRouteModal.value = true
}
</script>
<template>
  <a-list hoverable split>
    <template #header>
      <div class="flex justify-between">
        <span>服务列表</span>
        <a-button size="mini" type="primary" @click="editService({})">添加服务</a-button>
      </div>
    </template>
    <a-list-item v-for="service in serviceList" @click="emit('serviceSelected', service)">
      <div :class="{ 'font-italic': selectedService === service.id }"
           class="flex flex-col items-center justify-between cursor-pointer">
        <div class="w-100% flex justify-between">
          <span class="font-600 text-18px">{{ service.name }}</span>
          <div class="-mr-16px flex items-center">
            <a-dropdown-button size="mini" type="outline" @click="showDesensitiveDrawer = true">
              脱敏
              <template #content>
                <a-doption @click="editService(service)">编辑</a-doption>
                <a-doption @click="editServiceUserRoute(service)">用户信息拦截</a-doption>
              </template>
            </a-dropdown-button>
            <a-popconfirm content="确认删除该服务吗？" @ok="deleteServ(service)">
              <a-button size="mini" status="danger" type="text">
                <template #icon>
                  <icon-delete/>
                </template>
              </a-button>
            </a-popconfirm>
          </div>
        </div>
        <div class="text-12px mt-8px">{{ service.domain || '未配置' }}</div>
      </div>
    </a-list-item>
  </a-list>

  <!-- 添加服务 -->
  <a-modal v-model:visible="showServiceModal" :title="`${currentService.id ? '编辑' : '添加'}服务`" unmount-on-close
           @cancel="showServiceModal = false" @before-ok="handleServiceEditor">
    <a-form :model="currentService">
      <a-form-item field="name" label="服务名称">
        <a-input v-model:modelValue="currentService.name"/>
      </a-form-item>
      <a-form-item field="port" label="服务端口">
        <a-input-number v-model:modelValue="currentService.port" :max="65535" :min="1"/>
      </a-form-item>
      <a-form-item field="domain" label="匹配域名">
        <a-input v-model:modelValue="currentService.domain"/>
      </a-form-item>
    </a-form>
  </a-modal>

  <ServiceFieldDrawer v-if="showDesensitiveDrawer" :selected-service="selectedService"/>


  <!-- 用户信息拦截配置 -->
  <UserInfoRouteModal v-if="showUserInfoRouteModal" :service="userInfoRouteService"
                      @close="showUserInfoRouteModal = false"/>

</template>