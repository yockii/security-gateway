<script lang="ts" setup>
import { addRoute, deleteRoute, updateRoute } from '@/api/route';
import { saveRouteTarget } from '@/api/routeTarget';
import { getUpstreamList } from '@/api/upstream';
import { Route, RouteWithTarget } from '@/types/route';
import { Service } from '@/types/service';
import { Upstream } from '@/types/upstream';
import { Message, SelectOptionData } from '@arco-design/web-vue';
import { ref } from 'vue';
import RouteFieldDrawer from './RouteFieldDrawer.vue';

const emit = defineEmits(['routeSelected', 'routeUpdated'])
const props = defineProps<{
  selectedService: Service | undefined
  selectedRoute: string | undefined
  routes: RouteWithTarget[]
}>();


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
    currentRoute.value.serviceId = props.selectedService?.id
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
    emit('routeUpdated')
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
    const resp = await getUpstreamList({ name: value })
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
        emit('routeUpdated')
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

const showDesensitiveDrawer = ref(false)
const maskingRoute = ref<Route | undefined>()
const editMaskingRoute = (route: Route) => {
  maskingRoute.value = route
  showDesensitiveDrawer.value = true
}

const delRoute = async (route: RouteWithTarget) => {
  if (route.id) {
    try {
      const resp = await deleteRoute(route.id)
      if (resp.code === 0) {
        Message.success('删除成功')
        emit('routeUpdated')
      } else {
        Message.error('删除失败')
      }
    } catch (error) {
      console.log(error)
      Message.error('删除失败')
    }
  }
}
</script>

<template>
  <div class="flex justify-start items-center p-8px">
    <div class="font-600 text-18px">路由列表</div>
    <a-button class="ml-16px" size="mini" type="primary" @click="showRouteEditor({})">添加路由</a-button>
  </div>
  <div class="p-8px box-sizing">
    <a-space direction="vertical">
      <div v-if="selectedService && routes.length === 0">请添加路由</div>
      <div v-for="route in routes" class="flex items-center">
        <div class="p-8px border-solid border-1px border-#333 min-w-200px max-w-400px flex justify-between items-center"
          @click="emit('routeSelected', route)">
          <div class="font-italic font-600 text-18px mr-16px">{{ route.uri }}</div>

          <div class="-mr-8px flex items-center">
            <a-dropdown-button size="mini" type="outline" @click="editMaskingRoute(route)">
              脱敏
              <template #content>
                <a-doption @click="showRouteEditor(route)">编辑</a-doption>
              </template>
            </a-dropdown-button>
            <a-popconfirm content="确认删除该路由吗？" @ok="delRoute(route)">
              <a-button size="mini" status="danger" type="text">
                <template #icon>
                  <icon-delete />
                </template>
              </a-button>
            </a-popconfirm>
          </div>

        </div>
        <a-popover title="点击编辑目标">
          <icon-arrow-right class="text-32px mx-8px cursor-pointer hover:color-#165DFF hover:text-40px hover:mx-4px"
            @click="editTarget(route)" />
        </a-popover>
        <div class="p-8px border-solid border-1px border-#333 min-w-200px">
          <template v-if="editRouteTaret && currentRoute.id === route.id">
            <a-space>
              <!-- 选择目标上游 -->
              <a-select v-model="currentUpstream" :field-names="{ label: 'name' }" :filter-option="false"
                :loading="loadingUpstream" :style="{ width: '240px' }" allow-search placeholder="选择目标上游"
                @search="handleUpstreamSearch">
                <template #label="{ data }: { data: SelectOptionData }">
                  {{ (data?.value as Upstream).name }}
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
                    <icon-check />
                  </template>
                </a-button>
                <a-button type="outline" @click="cancelTargetUpstream">
                  <template #icon>
                    <icon-close />
                  </template>
                </a-button>
              </a-button-group>
            </a-space>
          </template>
          <template v-else>
            <div class="text-center font-600">{{ route.target?.name || '请配置' }}</div>
            <div class="text-center">{{ route.target?.targetUrl }}</div>
          </template>
        </div>
      </div>
    </a-space>
  </div>


  <!-- 添加路由 -->
  <a-modal v-model:visible="showRouteModal" :title="`${currentRoute.id ? '编辑' : '添加'}路由`" unmount-on-close
    @cancel="showRouteModal = false" @before-ok="handleRouteEditor">
    <a-form :model="currentRoute">
      <a-form-item field="uri" label="路由URI">
        <a-input v-model:modelValue="currentRoute.uri" />
      </a-form-item>
    </a-form>
  </a-modal>

  <RouteFieldDrawer v-if="showDesensitiveDrawer" :selected-route="maskingRoute"
    @close="showDesensitiveDrawer = false" />
</template>