<script lang="ts" setup>
import {addUser, deleteUser, getUserList} from '@/api/user';
import {User} from '@/types/user';
import {Message, PaginationProps, TableColumnData} from '@arco-design/web-vue';
import {onMounted, ref} from 'vue';
import moment from 'moment';
import {Response} from '@/types/common';

const conditionCollapsed = ref<boolean>(false);
const condition = ref<User>({
  secLevel: 0,
  page: 1,
  pageSize: 10,
})
const loading = ref<boolean>(false)
const list = ref<User[]>([])
const pagination = ref<PaginationProps>({
  total: 0,
  pageSize: 10,
})
const columns: TableColumnData[] = [
  {
    title: '用户名',
    dataIndex: 'username',
  },
  {
    title: '唯一标识',
    dataIndex: 'uniKey',
  },
  {
    title: '唯一标识Json',
    dataIndex: 'uniKeysJson',
    ellipsis: true,
    tooltip: true,
    width: 400
  },
  {
    title: '密级',
    dataIndex: 'secLevel',
  },
  {
    title: '操作',
    slotName: 'action',
  },
];

const getList = async () => {
  try {
    loading.value = true;
    const resp = await getUserList(condition.value);
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
const showUserModal = ref<boolean>(false);
const currentUser = ref<User>({});
const showEditor = (data: User) => {
  currentUser.value = data;
  showUserModal.value = true;
}
const saveUser = async (done: (closed: boolean) => void) => {
  let resp: Response<User> | undefined = undefined
  if (!currentUser.value.id) {
    // 新增，检查必填项
    if (!currentUser.value.username || (!currentUser.value.uniKey && !currentUser.value.uniKeysJson)) {
      Message.error('请填写完整信息');
      return;
    }
    try {
      resp = await addUser(currentUser.value);
    } catch (error) {
      console.error(error);
      Message.error('请求失败');
    }
  } else {
    // 编辑
    try {
      resp = await addUser(currentUser.value);
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
const readyToDelete = async (data: User) => {
  if (!data.id) {
    return;
  }
  try {
    const resp = await deleteUser(data.id);
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
          <span class="w-120px text-right">用户名：</span>
          <a-input v-model="condition.username" placeholder="名称"/>
        </a-grid-item>
        <a-grid-item class="flex items-center">
          <span class="w-160px text-right">通用唯一标识：</span>
          <a-input v-model="condition.uniKey" placeholder="唯一标识"/>
        </a-grid-item>
        <a-grid-item class="flex items-center">
          <span class="w-160px text-right">特定唯一标识：</span>
          <a-input v-model="condition.uniKeysJson" placeholder="特定唯一标识"/>
        </a-grid-item>
        <a-grid-item class="flex items-center">
          <span class="w-120px text-right">密级：</span>
          <a-select v-model="condition.secLevel">
            <a-option :value="0">所有</a-option>
            <a-option :value="1">一级</a-option>
            <a-option :value="2">二级</a-option>
            <a-option :value="3">三级</a-option>
            <a-option :value="4">四级</a-option>
          </a-select>
        </a-grid-item>
        <a-grid-item #="{ overflow }" class="flex justify-end" suffix>
          <a-button type="primary" @click="getList">查询</a-button>
          <a-button @click="conditionCollapsed = !conditionCollapsed">
            {{ overflow ? '收起' : '展开' }}
          </a-button>
          <a-button type="primary" @click="showEditor({})">新增</a-button>
        </a-grid-item>
      </a-grid>
      <a-table :columns="columns" :data="list" :loading="loading" :pagination="pagination"
               @page-change="pageChanged">
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
  <a-modal v-model:visible="showUserModal" title="编辑" unmount-on-close @cancel="showUserModal = false"
           @before-ok="saveUser">
    <a-form :model="currentUser">
      <a-form-item field="name" label="用户名">
        <a-input v-model="currentUser.username"/>
      </a-form-item>
      <a-form-item field="targetUrl" label="通用唯一标识">
        <a-input v-model="currentUser.uniKey"/>
      </a-form-item>
      <a-form-item field="targetUrl" label="特定唯一标识JSON">
        <a-textarea v-model="currentUser.uniKeysJson"/>
      </a-form-item>
      <a-form-item field="targetUrl" label="密级">
        <a-select v-model="currentUser.secLevel">
          <a-option :value="1">一级</a-option>
          <a-option :value="2">二级</a-option>
          <a-option :value="3">三级</a-option>
          <a-option :value="4">四级</a-option>
        </a-select>
      </a-form-item>
    </a-form>
  </a-modal>
</template>