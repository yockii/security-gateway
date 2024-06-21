<script lang="ts" setup>
import {getUpstreamListByRoute} from '@/api/upstream';
import {Route} from '@/types/route';
import {Upstream} from '@/types/upstream';
import {Message, PaginationProps} from '@arco-design/web-vue';
import {ref} from 'vue';
import RouteFieldDrawer from '../RouteFieldDrawer.vue';
import {Service} from '@/types/service';
import {addRoute, deleteRoute, updateRoute} from '@/api/route';

const props = defineProps<{
  service: Service | undefined
  routeList: Route[]
  paginationProps: PaginationProps
}>();
const emit = defineEmits(['pageChanged']);

const selectedRoute = ref<Route | undefined>(undefined);
const routeSelected = (route: Route) => {
  if (route && route.id === selectedRoute.value?.id) {
    return;
  }
  selectedRoute.value = route;
  route && getUpstreamByRoute();
}

const currentRoute = ref<Route>({});
const showRouteModal = ref(false)

const editRoute = (route: Route) => {
  // 复制一份，避免直接修改
  currentRoute.value = JSON.parse(JSON.stringify(route))
  showRouteModal.value = true
}
const handleRouteEdit = async (done: (close: boolean) => void) => {
  let resp = null
  if (!currentRoute.value.id) {
    // 新增
    // 检查路由URI是否为空
    if (!currentRoute.value.uri) {
      Message.warning('路由URI不能为空')
      return
    }
    currentRoute.value.serviceId = props.service?.id
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
    emit('pageChanged')
    done(true)
  } else {
    Message.error('操作失败')
  }
}

// 脱敏
const showDesensitiveDrawer = ref(false);
const routeMasking = (route: Route) => {
  currentRoute.value = route;
  showDesensitiveDrawer.value = true;
}

const delRoute = async (route: Route) => {
  if (!route.id) {
    return;
  }
  // 删除路由
  try {
    const resp = await deleteRoute(route.id);
    if (resp.code === 0) {
      Message.success('删除成功');
      emit('pageChanged');
    } else {
      Message.error('删除失败');
    }
  } catch (error) {
    console.log(error);
    Message.error('删除失败');
  }
}

// 上游
const upstreams = ref<Upstream[]>([]);
const upstreamPaginationProps = ref({
  defaultPageSize: 10,
  total: 0
});
const upstreamPage = ref<number>(1);
const getUpstreamByRoute = async () => {
  if (!selectedRoute.value || !selectedRoute.value.id) {
    return;
  }
  try {
    const resp = await getUpstreamListByRoute(selectedRoute.value.id);
    upstreams.value = resp.data?.items || [];
    upstreamPaginationProps.value.total = resp.data?.total || 0;
  } catch (error) {
    console.log(error);
  }
}
const upstreamPageChanged = (page: number) => {
  upstreamPage.value = page;
  getUpstreamByRoute();
}
</script>

<template>
  <a-split class="h-100% w-100%" default-size="320px" max="600px" min="280px">
    <template #first>
      <a-list :pagination-props="paginationProps" class="mx-8px" hoverable split
              @page-change="(current: number) => { emit('pageChanged', current) }">
        <template #header>
          <div class="flex justify-between">
            <span>路由列表</span>
            <a-button size="mini" type="primary" @click="editRoute({})">添加路由</a-button>
          </div>
        </template>
        <a-list-item v-for="route in routeList">
          <div class="flex items-center justify-between">
            <span :class="{ 'font-italic text-blue': selectedRoute && selectedRoute.id === route.id }"
                  class="flex-1 text-18px cursor-pointer" @click="routeSelected(route)">{{
                route.uri
              }}</span>
            <div class="-mr-16px flex items-center">
              <a-dropdown-button size="mini" type="outline" @click="routeMasking(route)">
                脱敏
                <template #content>
                  <a-doption @click="editRoute(route)">编辑</a-doption>
                </template>
              </a-dropdown-button>
              <a-popconfirm content="确认删除该路由吗？" @ok="delRoute(route)">
                <a-button size="mini" status="danger" type="text">
                  <template #icon>
                    <icon-delete/>
                  </template>
                </a-button>
              </a-popconfirm>
            </div>
          </div>
        </a-list-item>
      </a-list>
    </template>
    <template #second>
      <Upstreams :pagination-props="upstreamPaginationProps" :route="selectedRoute"
                 :upstreamList="routeList.length === 0 ? [] : upstreams" @pageChanged="upstreamPageChanged"/>
    </template>
  </a-split>

  <!-- 添加路由 -->
  <a-modal v-model:visible="showRouteModal" :title="`${currentRoute.id ? '编辑' : '添加'}路由`" unmount-on-close
           @cancel="showRouteModal = false" @before-ok="handleRouteEdit">
    <a-form :model="currentRoute">
      <a-form-item field="uri" label="路由URI">
        <a-input v-model:modelValue="currentRoute.uri"/>
      </a-form-item>
      <a-form-item field="loadBalance" label="负载均衡方式">
        <a-select v-model:model-value="currentRoute.loadBalance">
          <a-option :value="1">轮询</a-option>
          <a-option :value="2">权重</a-option>
          <a-option :value="3">IP哈希</a-option>
        </a-select>
      </a-form-item>
    </a-form>
  </a-modal>

  <RouteFieldDrawer v-if="showDesensitiveDrawer" :route="currentRoute" @close="showDesensitiveDrawer = false"/>
</template>