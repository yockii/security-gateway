<script lang="ts" setup>
import {addField, getFieldList, updateField} from '@/api/field';
import {addRoute, getRouteWithTargetList, updateRoute} from '@/api/route';
import {addService, getAllPorts, getServiceList, updateService} from '@/api/service';
import {Field} from '@/types/field';
import {RouteWithTarget} from '@/types/route';
import {Port, Service} from '@/types/service';
import {computed, onMounted, ref} from 'vue';
import {securityLevelToText} from '@/utils/security'
import {Message, SelectOptionData} from '@arco-design/web-vue';
import {Upstream} from '@/types/upstream';
import {getUpstreamList} from '@/api/upstream';
import {saveRouteTarget} from '@/api/routeTarget';

const ports = ref<Port[]>([])
const selectedPort = ref<number | undefined>(undefined)

const serviceList = ref<Service[]>([])
const selectedService = ref<Service | undefined>(undefined)

const getServiceInPort = async (port: number) => {
  if (port === 0) {
    return
  }
  if (selectedPort.value === port) {
    return
  }
  try {
    const resp = await getServiceList({port})
    serviceList.value = resp.data.items
    selectedPort.value = port
  } catch (error) {
    console.log(error)
  }
}

// 服务编辑
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
  } else {
    Message.error('操作失败')
  }
}
////////


const selectedRoute = ref<string | undefined>(undefined)

const routes = ref<RouteWithTarget[]>([])

const getRouteList = async (service: Service | undefined) => {
  if (!service) return
  if (selectedService.value === service) {
    return
  }
  try {
    const resp = await getRouteWithTargetList({serviceId: service.id})
    routes.value = resp.data.items
    selectedService.value = service
  } catch (error) {
    console.log(error)
  }
}

// 路由编辑
const currentRoute = ref<RouteWithTarget>({})
const showRouteModal = ref(false)
const showRouteEditor = (route: RouteWithTarget) => {
  // 复制一份，避免直接修改
  currentRoute.value = JSON.parse(JSON.stringify(route))
  showRouteModal.value = true
}
const handleRouteEditor = async (done: (close: boolean) => void) => {
  let resp = null
  if (!currentRoute.value.id) {
    // 新增
    // 检查路由URI是否为空
    if (!currentRoute.value.uri) {
      Message.warning('路由URI不能为空')
      return
    }
    currentRoute.value.serviceId = selectedService.value?.id
    try {
      resp = await addRoute(currentRoute.value)
    } catch (error) {
      console.log(error)
      Message.error('新增路由失败')
      return
    }
  } else {
    try {
      resp = await updateRoute(currentRoute.value)
    } catch (error) {
      console.log(error)
      Message.error('更新路由失败')
      return
    }
  }
  if (resp.code === 0) {
    Message.success('操作成功')
    getRouteList(selectedService.value)
    done(true)
  } else {
    Message.error('操作失败')
  }
}

// 目标上游编辑
const currentUpstream = ref<Upstream | undefined>()
const editRouteTaret = ref(false)
const loadingUpstream = ref(false)
const upstreamList = ref<Upstream[]>([])
const editTarget = (route: RouteWithTarget) => {
  currentRoute.value = JSON.parse(JSON.stringify(route))
  currentUpstream.value = currentRoute.value.target
  if (upstreamList.value.length === 0) {
    handleUpstreamSearch('')
  }
  editRouteTaret.value = true
}
const handleUpstreamSearch = async (value: string) => {
  loadingUpstream.value = true
  try {
    const resp = await getUpstreamList({name: value})
    upstreamList.value = resp.data.items
  } catch (error) {
    console.log(error)
  } finally {
    loadingUpstream.value = false
  }
}
const confirmTargetUpstream = async () => {
  if (!currentUpstream.value) {
    return
  }
  if (currentRoute.value.id && currentUpstream.value.id) {
    try {
      const resp = await saveRouteTarget({
        routeId: currentRoute.value.id,
        upstreamId: currentUpstream.value.id
      })
      if (resp.code === 0) {
        Message.success('操作成功')
        editRouteTaret.value = false
        currentUpstream.value = undefined
        currentRoute.value = {}
        getRouteList(selectedService.value)
      } else {
        Message.error('操作失败')
      }
    } catch (error) {
      console.log(error)
      Message.error('操作失败')
    }
  }
}
const cancelTargetUpstream = () => {
  editRouteTaret.value = false
  currentUpstream.value = undefined
  currentRoute.value = {}
}


// 脱敏配置
const showDesensitive = ref(false)
const desensitiveTitle = computed(() => {
  return selectedService.value?.name + '的脱敏配置'
})
const desensitiveFieldList = ref<Field[]>([])
const fieldsTotal = ref(0)
const fieldCondition = ref<Field>({
  page: 1,
  pageSize: 5,
})
const getDesensitiveFieldList = async () => {
  return await getFieldList(fieldCondition.value)
}
const fieldPageChanegd = async (current: number) => {
  fieldCondition.value.page = current
  try {
    const resp = await getDesensitiveFieldList()
    desensitiveFieldList.value = resp.data.items
    fieldsTotal.value = resp.data.total
  } catch (error) {
    console.log(error)
  }
}
const showDesensitiveDrawer = async () => {
  if (!selectedService.value) {
    return
  }
  fieldCondition.value.serviceId = selectedService.value.id
  try {
    const resp = await getDesensitiveFieldList()
    desensitiveFieldList.value = resp.data.items
    fieldsTotal.value = resp.data.total
    showDesensitive.value = true
  } catch (error) {
    console.log(error)
  }
}

// 字段编辑
const currentField = ref<Field>({})
const showFieldEditModal = ref(false)
const editField = (field: Field) => {
  currentField.value = field
  showFieldEditModal.value = true
}
const handleFieldEditor = async (done: (closed: boolean) => void) => {
  let resp = null
  if (!currentField.value.id) {
    // 新增
    // 检查字段名称是否为空
    if (!currentField.value.fieldName) {
      Message.warning('字段名称不能为空')
      return
    }
    try {
      resp = await addField(currentField.value)
    } catch (error) {
      console.log(error)
      Message.error('新增字段失败')
      return
    }
  } else {
    try {
      resp = await updateField(currentField.value)
    } catch (error) {
      console.log(error)
      Message.error('更新字段失败')
      return
    }
  }
  if (resp.code === 0) {
    Message.success('操作成功')
    fieldPageChanegd(fieldCondition.value.page || 1)
    done(true)
  } else {
    Message.error('操作失败')
  }

}

onMounted(() => {
  getAllPorts().then((response) => {
    ports.value = response.data;
  });
});
</script>

<template>
  <a-layout class="p-16px">
    <!-- 端口列表 -->
    <a-layout-sider style="width: 120px;">
      <a-list hoverable split>
        <template #header>
          <div>端口列表</div>
        </template>
        <a-list-item v-for="port in ports" @click="getServiceInPort(port.port)">
          <div :class="{ 'font-600 font-italic': selectedPort === port.port }"
               class="flex items-center justify-between cursor-pointer">
            <span>{{ port.port }}</span>
            <div v-if="port.inUse" class="w-10px h-10px b-rd-50% bg-green"></div>
            <div v-else class="w-10px h-10px b-rd-50% bg-gray"></div>
          </div>
        </a-list-item>
      </a-list>
    </a-layout-sider>
    <!-- 服务列表 -->
    <a-layout-sider style="min-width: 240px; margin-left: 1px;">
      <a-list hoverable split>
        <template #header>
          <div class="flex justify-between">
            <span>服务列表</span>
            <a-button size="mini" type="primary" @click="editService({})">添加服务</a-button>
          </div>
        </template>
        <a-list-item v-for="service in serviceList" @click="getRouteList(service)">
          <div :class="{ 'font-italic': selectedService === service.id }"
               class="flex flex-col items-center justify-between cursor-pointer">
            <div class="w-100% flex justify-between">
              <span class="font-600 text-18px">{{ service.name }}</span>
              <a-button-group size="mini">
                <a-button type="primary" @click="editService(service)">编辑</a-button>
                <a-button type="outline" @click="showDesensitiveDrawer">脱敏</a-button>
              </a-button-group>
            </div>
            <div class="text-12px mt-8px">{{ service.domain || '未配置' }}</div>
          </div>
        </a-list-item>
      </a-list>
    </a-layout-sider>
    <!-- 路由列表 -->
    <a-layout-content>
      <div class="flex justify-start items-center p-8px">
        <div class="font-600 text-18px">路由列表</div>
        <a-button class="ml-16px" size="mini" type="primary" @click="showRouteEditor({})">添加路由</a-button>
      </div>
      <div class="p-8px box-sizing">
        <a-space direction="vertical">
          <div v-for="route in routes" class="flex items-center">
            <div class="p-8px border-solid border-1px border-#333 min-w-200px max-w-400px flex justify-between items-center"
                 @click="selectedRoute = route.uri">
              <div class="font-italic font-600 text-18px mr-16px">{{ route.uri }}</div>

              <a-button size="mini" type="primary" @click="showRouteEditor(route)">
                <template #icon>
                  <icon-edit/>
                </template>
              </a-button>
            </div>
            <a-popover title="点击编辑目标">
              <icon-arrow-right class="text-32px mx-8px cursor-pointer hover:color-#165DFF hover:text-40px hover:mx-4px"
                                @click="editTarget(route)"/>
            </a-popover>
            <div class="p-8px border-solid border-1px border-#333 min-w-200px">
              <template v-if="editRouteTaret && currentRoute.id === route.id">
                <a-space>
                  <!-- 选择目标上游 -->
                  <a-select v-model="currentUpstream" :field-names="{ label: 'name' }"
                            :filter-option="false" :loading="loadingUpstream" :style="{ width: '240px' }"
                            allow-search placeholder="选择目标上游" @search="handleUpstreamSearch">
                    <template #label="{ data }: { data: SelectOptionData }">
                      {{ (data?.value as Upstream).name || 'aaa' }}
                    </template>
                    <a-option v-for="item of upstreamList" :value="item">
                      <div>
                        {{ item.name }}
                      </div>
                      <div>{{ item.targetUrl }}</div>
                    </a-option>
                  </a-select>
                  <a-button-group>
                    <a-button type="primary" @click="confirmTargetUpstream">
                      <template #icon>
                        <icon-check/>
                      </template>
                    </a-button>
                    <a-button type="outline" @click="cancelTargetUpstream">
                      <template #icon>
                        <icon-close/>
                      </template>
                    </a-button>
                  </a-button-group>
                </a-space>
              </template>
              <template v-else>
                <div class="text-center font-600">{{ route.target?.name || '请配置' }}</div>
                <div>{{ route.target?.targetUrl }}</div>
              </template>
            </div>
          </div>
        </a-space>
      </div>
    </a-layout-content>
  </a-layout>

  <!-- 添加服务 -->
  <a-modal v-model:visible="showServiceModal" :title="`${currentService.id ? '编辑' : '添加'}服务`"
           unmount-on-close @cancel="showServiceModal = false" @before-ok="handleServiceEditor">
    <a-form :model="currentService">
      <a-form-item field="name" label="服务名称">
        <a-input v-model:modelValue="currentService.name"/>
      </a-form-item>
      <a-form-item field="port" label="服务端口">
        <a-input-number v-model:modelValue="currentService.port" :max="65535" :min="1"/>
      </a-form-item>
      <a-form-item field="domain" label="服务域名">
        <a-input v-model:modelValue="currentService.domain"/>
      </a-form-item>
    </a-form>
  </a-modal>

  <!-- 添加路由 -->
  <a-modal v-model:visible="showRouteModal" :title="`${currentRoute.id ? '编辑' : '添加'}路由`"
           unmount-on-close @cancel="showRouteModal = false" @before-ok="handleRouteEditor">
    <a-form :model="currentRoute">
      <a-form-item field="uri" label="路由URI">
        <a-input v-model:modelValue="currentRoute.uri"/>
      </a-form-item>
    </a-form>
  </a-modal>

  <!-- 脱敏配置抽屉 -->
  <a-drawer :visible="showDesensitive" :width="520" @cancel="showDesensitive = false">
    <template #title>
      <a-space>
        <span>{{ desensitiveTitle }}</span>
        <a-button size="mini" type="primary" @click="editField({})">添加字段</a-button>
      </a-space>
    </template>
    <template #footer>
      <div class="flex justify-end">
        <a-pagination :current="fieldCondition.page" :page-size="fieldCondition.pageSize" :total="fieldsTotal"
                      size="mini" @change="fieldPageChanegd"/>
      </div>
    </template>
    <a-space direction="vertical" fill>
      <template v-for="field in desensitiveFieldList">
        <a-card :title="field.fieldName" size="small">
          <template #extra>
            <a-button size="mini" type="primary" @click="editField(field)">编辑</a-button>
          </template>
          <div>
            <div><span>一级密级:</span><span>{{ securityLevelToText(field.level1) }}</span></div>
            <div><span>二级密级:</span><span>{{ securityLevelToText(field.level2) }}</span></div>
            <div><span>三级密级:</span><span>{{ securityLevelToText(field.level3) }}</span></div>
            <div><span>四级密级:</span><span>{{ securityLevelToText(field.level4) }}</span></div>
          </div>
        </a-card>
      </template>
    </a-space>
  </a-drawer>

  <!-- 脱敏字段编辑弹框 -->
  <a-modal v-model:visible="showFieldEditModal" :title="currentField.fieldName" unmount-on-close
           @cancel="showFieldEditModal = false" @before-ok="handleFieldEditor">
    <a-form :model="currentField">
      <a-form-item field="fieldName" label="字段名称">
        <a-input v-model:modelValue="currentField.fieldName" :disabled="!!currentField.id"/>
      </a-form-item>
      <a-form-item field="comment" label="字段描述">
        <a-input v-model:modelValue="currentField.comment"/>
      </a-form-item>
      <a-form-item field="level1" label="一级密级">
        <SecurityLevel v-model:modelValue="currentField.level1"/>
      </a-form-item>
      <a-form-item field="level2" label="二级密级">
        <SecurityLevel v-model:modelValue="currentField.level2"/>
      </a-form-item>
      <a-form-item field="level3" label="三级密级">
        <SecurityLevel v-model:modelValue="currentField.level3"/>
      </a-form-item>
      <a-form-item field="level4" label="四级密级">
        <SecurityLevel v-model:modelValue="currentField.level4"/>
      </a-form-item>
    </a-form>
  </a-modal>
</template>