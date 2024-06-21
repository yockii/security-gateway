<script lang="ts" setup>
import {Route} from '@/types/route';
import {Service} from '@/types/service';
import {reactive, ref} from 'vue';
import Routes from './Routes.vue';
import {Message, PaginationProps} from '@arco-design/web-vue';
import {getRouteList} from '@/api/route';
import {addService, deleteService, updateService} from '@/api/service';
import ServiceFieldDrawer from '../ServiceFieldDrawer.vue';

const props = defineProps<{
  serviceList: Service[]
  paginationProps: PaginationProps
}>();
const emit = defineEmits(['pageChanged', 'portChanged']);

const selectedService = ref<Service | undefined>(undefined);
const serviceSelected = (service: Service) => {
  if (service && service.id === selectedService.value?.id) {
    return;
  }
  selectedService.value = service;
  service && getRoutesInService();
}

// 增改服务
const currentService = ref<Service>({});
const showServiceModal = ref(false);
const editService = (service: Service) => {
  // 复制一份，避免直接修改
  currentService.value = JSON.parse(JSON.stringify(service))
  showServiceModal.value = true
}
const handleServiceEdit = async (done: (closed: boolean) => void) => {
  let resp = null
  let portChanged = false
  if (!currentService.value.id) {
    // 新增
    // 检查服务名称是否为空
    if (!currentService.value.name || !currentService.value.port) {
      Message.warning('服务名称及端口不能为空')
      return
    }
    try {
      resp = await addService(currentService.value)
      portChanged = true
    } catch (error) {
      console.log(error)
      Message.error('新增服务失败')
      return
    }
  } else {
    try {
      resp = await updateService(currentService.value)

      portChanged = currentService.value.port !== props.serviceList.find((item) => item.id === currentService.value.id)?.port

    } catch (error) {
      console.log(error)
      Message.error('更新服务失败')
      return
    }
  }
  if (resp.code === 0) {
    Message.success('操作成功')
    done(true)
    if (portChanged) {
      emit('portChanged')
    }
    emit('pageChanged')
  } else {
    Message.error('操作失败')
  }
}

// 脱敏
const showDesensitiveDrawer = ref(false);
const serviceMasking = (service: Service) => {
  currentService.value = service
  showDesensitiveDrawer.value = true
}

// 用户信息拦截配置
const showUserInfoRouteModal = ref(false)
const editServiceUserRoute = (service: Service) => {
  currentService.value = service
  showUserInfoRouteModal.value = true
}

const delService = async (service: Service) => {
  try {
    const resp = await deleteService(service.id!)
    if (resp.code === 0) {
      Message.success('删除成功')
      emit('pageChanged')
      routes.value = []
    } else {
      Message.error('删除失败')
    }
  } catch (error) {
    console.log(error)
    Message.error('删除失败')
  }
}


// 路由
const routes = ref<Route[]>([]);
const routePaginationProps = reactive({
  defaultPageSize: 10,
  total: 0
})
const routePage = ref<number>(1);
const getRoutesInService = async () => {
  if (!selectedService.value) {
    return;
  }
  try {
    const resp = await getRouteList({serviceId: selectedService.value.id, page: routePage.value});
    routes.value = resp.data?.items || [];
    routePaginationProps.total = resp.data?.total || 0;
  } catch (error) {
    console.log(error);
  }
}
const routePageChanged = (page: number) => {
  routePage.value = page;
  getRoutesInService();
}
</script>

<template>
  <a-split class="h-100% w-100%" default-size="240px" max="400px" min="200px">
    <template #first>
      <a-list :pagination-props="paginationProps" class="mx-8px" hoverable split
              @page-change="(current: number) => { emit('pageChanged', current) }">
        <template #header>
          <div class="flex justify-between">
            <span>服务列表</span>
            <a-button size="mini" type="primary" @click="editService({})">添加服务</a-button>
          </div>
        </template>
        <a-list-item v-for="service in serviceList" @click="serviceSelected(service)">
          <div class="flex flex-col items-center justify-between cursor-pointer">
            <div class="w-100% flex justify-between">
              <span :class="{ 'font-italic font-600 text-blue': selectedService && selectedService.id === service.id }"
                    class="text-18px">{{
                  service.name
                }}</span>
              <div class="-mr-16px flex items-center" @click.stop>
                <a-dropdown-button size="mini" type="outline" @click.stop="serviceMasking(service)">
                  脱敏
                  <template #content>
                    <a-doption @click="editService(service)">编辑</a-doption>
                    <a-doption @click="editServiceUserRoute(service)">用户信息拦截</a-doption>
                  </template>
                </a-dropdown-button>
                <a-popconfirm content="确认删除该服务吗？" @ok="delService(service)">
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
    </template>
    <template #second>
      <Routes :pagination-props="routePaginationProps" :route-list="serviceList.length == 0 ? [] : routes"
              :service="selectedService" @page-changed="routePageChanged"/>
    </template>
  </a-split>


  <!-- 增改服务 -->
  <a-modal v-model:visible="showServiceModal" :title="`${currentService.id ? '编辑' : '添加'}服务`" unmount-on-close
           @cancel="showServiceModal = false" @before-ok="handleServiceEdit">
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

  <!-- 脱敏抽屉 -->
  <ServiceFieldDrawer v-if="showDesensitiveDrawer" :service="currentService" @closed="showDesensitiveDrawer = false"/>

  <!-- 用户信息拦截配置 -->
  <UserInfoRouteModal v-if="showUserInfoRouteModal" :service="currentService" @close="showUserInfoRouteModal = false"/>

</template>