<script lang="ts" setup>
import {Route} from '@/types/route';
import {Service} from '@/types/service';
import {Certificate} from '@/types/certificate';
import {reactive, ref} from 'vue';
import Routes from './Routes.vue';
import {Message, PaginationProps} from '@arco-design/web-vue';
import {getRouteList} from '@/api/route';
import {addService, deleteService, updateService, updateServiceCert} from '@/api/service';
import {listByDomain} from '@/api/certificate';
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


// 配置证书
const showCertificateModal = ref(false);
const certList = ref<Certificate[]>([]);
const editServiceCertficate = async (service: Service) => {
  if (!service.domain) {
    Message.warning('请先配置服务域名');
    return;
  }
  currentService.value = service;
  try {
    const resp = await listByDomain(service.domain);
    if (resp.code === 0) {
      certList.value = resp.data || [];
    } else {
      Message.error('获取证书信息失败');
    }
    showCertificateModal.value = true;
  } catch (error) {
    console.log(error);
    Message.error('获取证书信息失败');
  }
}
const confirmServiceCert = async () => {
  if (!currentService.value.id) {
    Message.warning('未指定服务');
    return;
  }
  // 更新服务
  try {
    const resp = await updateServiceCert(currentService.value.id, currentService.value.certificateId || "");
    if (resp.code === 0) {
      Message.success('操作成功');
      showCertificateModal.value = false;
    } else {
      Message.error('操作失败');
    }
  } catch (error) {
    console.log(error);
    Message.error('操作失败');
  }
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
                    <a-doption :disabled="!service.domain" @click="editServiceCertficate(service)">证书配置</a-doption>
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

  <!-- 证书配置抽屉 -->
  <a-drawer :visible="showCertificateModal" :width="550" @cancel="showCertificateModal = false"
            @close="showCertificateModal = false" @ok="confirmServiceCert">
    <template #title>
      <div class="w-500px flex justify-between">
        <span>证书管理</span>
        <a-button size="mini" status="danger" type="primary"
                  @click="currentService.certificateId = '0'">删除绑定的证书
        </a-button>
      </div>
    </template>
    <a-radio-group v-model:model-value="currentService.certificateId" :default-value="currentService.certificateId">
      <template v-for="cert in certList" :key="cert.id">
        <a-radio :value="cert.id">
          <template #radio="{ checked }">
            <a-space :class="{ 'bg-#e8f3ff  b-#165DFF': checked }" class="py-8px px-16px b-1px b-solid b-gray b-rd-4px w-200px relative"
                     direction="vertical">
              <div :class="{ 'text-#165dff': checked }" class="text-18px font-600">{{ cert.certName }}</div>
              <div class="text-14px">{{ cert.serveDomain }}</div>
              <div class="text-12px">{{ cert.certDesc }}</div>

              <div
                  class="absolute right-16px top-16px w-14px h-14px inline-flex items-center justify-center b-rd-100% b-1px b-solid b-gray">
                <div :class="{ 'bg-#165dff': checked }" class="w-8px h-8px b-rd-100%"></div>
              </div>
            </a-space>
          </template>
        </a-radio>
      </template>
    </a-radio-group>
  </a-drawer>
</template>