<script lang="ts" setup>
import { addUpstream, deleteUpstream, getUpstreamList } from '@/api/upstream';
import { Upstream } from '@/types/upstream';
import { Message, PaginationProps, TableColumnData } from '@arco-design/web-vue';
import { onMounted, ref } from 'vue';
import moment from 'moment';

const conditionCollapsed = ref<boolean>(false);
const condition = ref<Upstream>({
  page: 1,
  pageSize: 10,
})
const loading = ref<boolean>(false)
const list = ref<Upstream[]>([])
const pagination = ref<PaginationProps>({
  total: 0,
  pageSize: 10,
})
const columns: TableColumnData[] = [
  {
    title: '名称',
    dataIndex: 'name',
  },
  {
    title: '目标地址',
    dataIndex: 'targetUrl',
  },
  {
    title: '创建时间',
    dataIndex: 'createTime',
    slotName: 'time',
  },
  {
    title: '操作',
    slotName: 'action',
  },
];

const getList = async () => {
  try {
    loading.value = true;
    const resp = await getUpstreamList(condition.value);
    if (resp.code === 0) {
      list.value = resp.data?.items || [];
      pagination.value.total = resp.data?.total || 0;
    } else {
      console.error(resp.msg);
      Message.error(resp.msg);
    }
  } catch (error) {
    console.error(error);
    Message.error('请求失败');
  } finally {
    loading.value = false;
  }
}

// 表格分页处理
const pageChanged = (page: number) => {
  condition.value.page = page;
  getList();
}

// 编辑
const showUpstreamModal = ref<boolean>(false);
const currentUpstream = ref<Upstream>({});
const showEditor = (data: Upstream) => {
  currentUpstream.value = data;
  showUpstreamModal.value = true;
}
const saveUpstream = async (done: (closed: boolean) => void) => {
  let resp = null
  if (!currentUpstream.value.id) {
    // 新增，检查必填项
    if (!currentUpstream.value.name || !currentUpstream.value.targetUrl) {
      Message.error('请填写完整信息');
      return;
    }
    try {
      resp = await addUpstream(currentUpstream.value);
    } catch (error) {
      console.error(error);
      Message.error('请求失败');
    }
  } else {
    // 编辑
    try {
      resp = await addUpstream(currentUpstream.value);
    } catch (error) {
      console.error(error);
      Message.error('请求失败');
    }
  }
  if (resp && resp.code === 0) {
    Message.success('保存成功');
    getList();
  } else {
    console.error(resp?.msg);
    Message.error(resp?.msg || '请求失败');
  }
  done(true)
}

// 删除
const readyToDelete = async (data: Upstream) => {
  if (!data.id) {
    return;
  }
  try {
    const resp = await deleteUpstream(data.id);
    if (resp.code === 0) {
      Message.success('删除成功');
      getList();
    } else {
      console.error(resp.msg);
      Message.error(resp.msg);
    }
  } catch (error) {
    console.error(error);
    Message.error('请求失败');
  } finally {
    getList();
  }
}

onMounted(() => {
  getList()
})
</script>

<template>
  <a-layout-content class="p-16px">
    <a-space direction="vertical" size="large" style="width: 100%;">
      <a-grid :col-gap="16" :collapsed="conditionCollapsed" :cols="{ xs: 1, sm: 2, md: 3, lg: 4, xl: 5, xxl: 6 }"
        :row-gap="8">
        <a-grid-item class="flex items-center">
          <span class="w-120px text-right">名称：</span>
          <a-input v-model="condition.name" placeholder="名称" />
        </a-grid-item>
        <a-grid-item class="flex items-center">
          <span class="w-120px text-right">目标地址：</span>
          <a-input v-model="condition.targetUrl" placeholder="目标地址" />
        </a-grid-item>
        <a-grid-item #="{ overflow }" class="flex justify-end" suffix>
          <a-button type="primary" @click="getList">查询</a-button>
          <a-button @click="conditionCollapsed = !conditionCollapsed">
            {{ overflow ? '收起' : '展开' }}
          </a-button>
          <a-button type="primary" @click="showEditor({})">新增</a-button>
        </a-grid-item>
      </a-grid>
      <a-table :columns="columns" :data="list" :loading="loading" :pagination="pagination" @page-change="pageChanged">
        <template #time="{ record }">
          {{ moment(record.createTime).format('YYYY-MM-DD HH:mm:ss') }}
        </template>
        <template #action="{ record }">
          <a-button-group>
            <a-button type="primary" @click="showEditor(record)">编辑</a-button>
            <a-popconfirm content="确认删除吗？" @ok="readyToDelete(record)">
              <a-button status="danger" type="outline">删除</a-button>
            </a-popconfirm>
          </a-button-group>
        </template>
      </a-table>
    </a-space>
  </a-layout-content>

  <!-- 编辑弹窗 -->
  <a-modal v-model:visible="showUpstreamModal" title="编辑" unmount-on-close @cancel="showUpstreamModal = false"
    @before-ok="saveUpstream">
    <a-form :model="currentUpstream">
      <a-form-item field="name" label="名称">
        <a-input v-model="currentUpstream.name" />
      </a-form-item>
      <a-form-item field="targetUrl" label="目标地址">
        <a-input v-model="currentUpstream.targetUrl" />
      </a-form-item>
    </a-form>
  </a-modal>
</template>