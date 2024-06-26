<script lang="ts" setup>
import {deleteRouteTarget, saveRouteTarget} from '@/api/routeTarget';
import {getUpstreamList} from '@/api/upstream';
import {Route} from '@/types/route';
import {TargetWithUpstream, Upstream} from '@/types/upstream'
import {Message, PaginationProps} from '@arco-design/web-vue'
import {computed, ref} from 'vue';

const props = defineProps<{
  route: Route
  upstreamList: TargetWithUpstream[]
  paginationProps: PaginationProps
}>();
const emit = defineEmits(['pageChanged'])
const loadBalance = computed(() => {
  switch (props.route?.loadBalance) {
    case 1:
      return '轮询'
    case 2:
      return '权重'
    case 3:
      return 'IP哈希'
    default:
      return '未知'
  }
})

// 上游目标编辑
const showTargetUpstreamModal = ref(false)
const targetUpstream = ref<TargetWithUpstream>({
  routeId: props.route?.id || ''
})
const loadingUpstream = ref(false)
const searchedUpstreamList = ref<Upstream[]>([])
const handleUpstreamSearch = async (value: string) => {
  loadingUpstream.value = true
  try {
    const resp = await getUpstreamList({name: value})
    searchedUpstreamList.value = resp.data?.items || []
  } catch (error) {
    console.log(error)
  } finally {
    loadingUpstream.value = false
  }
}
const editTargetUpstream = (tu: TargetWithUpstream) => {
  targetUpstream.value = {
    id: tu.id || '',
    routeId: props.route?.id || '',
    upstreamId: tu.upstreamId,
    weight: tu.weight
  }
  handleUpstreamSearch('')
  showTargetUpstreamModal.value = true
}

const cancelEdit = () => {
  targetUpstream.value = {
    routeId: props.route?.id || ''
  }
  showTargetUpstreamModal.value = false
}

const confirmTargetUpstream = async (done: (close: boolean) => void) => {
  if (!targetUpstream.value || !targetUpstream.value.upstreamId) {
    return
  }
  if (!targetUpstream.value.weight) {
    Message.warning('权重不能为空')
    return
  }
  if (!targetUpstream.value.routeId) {
    targetUpstream.value.routeId = props.route?.id || ''
  }
  if (props.route?.id) {
    try {
      const resp = await saveRouteTarget(targetUpstream.value)
      if (resp.code === 0) {
        Message.success('操作成功')
        targetUpstream.value = {
          routeId: props.route?.id || ''
        }
        showTargetUpstreamModal.value = false
        emit('pageChanged')
        done(true)
      } else {
        Message.error('操作失败')
      }
    } catch (error) {
      console.log(error)
      Message.error('操作失败')
    }
  }
}

// 删除
const delTargetUpstream = async (tu: TargetWithUpstream) => {
  if (!tu.id) {
    return
  }
  try {
    const resp = await deleteRouteTarget(tu.id)
    if (resp.code === 0) {
      Message.success('删除成功')
      emit('pageChanged')
    } else {
      Message.error('删除失败')
    }
  } catch (error) {
    console.log(error)
    Message.error('删除失败')
  }
}
</script>

<template>
  <a-list :pagination-props="paginationProps" class="mx-8px" hoverable split
          @page-change="(current: number) => { emit('pageChanged', current) }">
    <template #header>
      <div class="flex justify-between">
        <span>上游目标列表 [{{ loadBalance }}]</span>
        <a-button size="mini" type="primary" @click="editTargetUpstream({})">添加上游目标</a-button>
      </div>
    </template>
    <a-list-item v-for="tu in upstreamList" :key="tu.id">
      <a-list-item-meta :description="tu.upstream?.targetUrl" :title="tu.upstream?.name">
        <template #avatar>
          <div class="bg-blue px-12px py-8px text-white text-12px flex items-center b-rd-md">
            <span>权重：</span>
            <span class="text-20px font-600">{{ tu.weight }}</span>
          </div>
        </template>
      </a-list-item-meta>
      <template #actions>
        <a-button size="mini" type="text" @click="editTargetUpstream(tu)">
          <icon-edit/>
        </a-button>
        <a-popconfirm content="确认删除该上游目标吗？" @ok="delTargetUpstream(tu)">
          <a-button size="mini" status="danger" type="text">
            <icon-delete/>
          </a-button>
        </a-popconfirm>

      </template>
    </a-list-item>
  </a-list>


  <!-- 添加上游目标 -->
  <a-modal v-model:visible="showTargetUpstreamModal" :title="`${targetUpstream.upstreamId ? '编辑' : '添加'}上游目标`"
           unmount-on-close @cancel="cancelEdit" @before-ok="confirmTargetUpstream">
    <a-form :model="targetUpstream">
      <a-form-item field="loadBalance" label="负载均衡方式">
        <!-- 选择目标上游 -->
        <a-select v-model:model-value="targetUpstream.upstreamId" :loading="loadingUpstream" allow-search
                  placeholder="选择目标上游" @search="handleUpstreamSearch">
          <template #label>
            {{ searchedUpstreamList.find(item => item.id === targetUpstream.upstreamId)?.name || '选择目标上游' }}
          </template>
          <a-option v-for="item of searchedUpstreamList" :value="item.id">
            <div>
              {{ item.name }}
            </div>
            <div class="flex items-center">
              <span v-if="item.healthCheckUrl" :class="{ 'bg-red': item.status === 2, 'bg-green': item.status === 1, 'bg-orange': item.status === 0 }"
                    class="mr-8px w-8px h-8px b-rd-100% "></span>
              <span v-else class="mr-8px w-8px h-8px b-rd-100% bg-gray"></span>
              <span>{{ item.targetUrl }}</span>
            </div>
          </a-option>
        </a-select>
      </a-form-item>
      <!-- 权重 -->
      <a-form-item field="weight" label="权重">
        <a-input-number v-model="targetUpstream.weight" :max="100" :min="0"/>
      </a-form-item>
    </a-form>
  </a-modal>
</template>