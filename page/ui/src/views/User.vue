<script lang="ts" setup>
import {addUser, deleteUser, getUserList} from '@/api/user';
import {User} from '@/types/user';
import {Message, PaginationProps, TableColumnData} from '@arco-design/web-vue';
import {computed, onMounted, ref} from 'vue';
import moment from 'moment';
import {Response} from '@/types/common';
import {UserServiceLevel} from '@/types/userServiceLevel';
import {addUserServiceLevel, getUserServiceLevelListWithService} from '@/api/userServiceLevel';
import {Service} from '@/types/service';
import {getServiceList} from '@/api/service';

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

// 服务密级配置
const userServiceLevelList = ref<UserServiceLevel[]>([])
const userServiceLevelTotal = ref<number>(0)
const userServiceLevelUser = ref<User>({})
const showServiceSecretLevelDrawer = ref<boolean>(false)
const showServiceSecretLevelEditor = (data: User) => {
  userServiceLevelUser.value = data;
  userServiceLevelList.value = []
  showServiceSecretLevelDrawer.value = true;
  getListUserServiceLevel();
}
const getListUserServiceLevel = async () => {
  try {
    const resp = await getUserServiceLevelListWithService({userId: userServiceLevelUser.value.id});
    if (resp.code === 0) {
      userServiceLevelList.value = resp.data?.items || [];
      userServiceLevelTotal.value = resp.data?.total || 0;
    } else {
      console.error(resp.msg);
      Message.error(resp.msg);
    }
  } catch (error) {
    console.error(error);
    Message.error('请求失败');
  }
}
// 编辑服务密级
const showServiceLevelModal = ref<boolean>(false)
const currentServiceLevel = ref<UserServiceLevel>({})
const addNewServiceLevel = () => {
  currentServiceLevel.value = {
    userId: userServiceLevelUser.value.id,
    secLevel: 1,
  }
  showServiceLevelModal.value = true
}
// 搜索服务
const serviceList = ref<Service[]>([])
const canSelecteServiceList = computed(() => {
  const result = serviceList.value.concat()
  for (let i = result.length - 1; i >= 0; i--) {
    for (let j = 0; j < userServiceLevelList.value.length; j++) {
      if (result[i].id === userServiceLevelList.value[j].serviceId) {
        result.splice(i, 1)
        break
      }
    }
  }
  return result
})
const serviceLoading = ref<boolean>(false)
const handleServiceSearch = async (value: string) => {
  serviceLoading.value = true
  try {
    const resp = await getServiceList({name: value})
    if (resp.code === 0) {
      serviceList.value = resp.data?.items || []
    } else {
      console.error(resp.msg)
      Message.error(resp.msg)
    }
  } catch (error) {
    console.error(error)
    Message.error('请求失败')
  } finally {
    serviceLoading.value = false
  }
}
// 保存
const saveServiceLevel = async () => {
  let resp: Response<UserServiceLevel> | undefined = undefined
  if (currentServiceLevel.value.id) {
    // 编辑
    try {
      resp = await addUserServiceLevel(currentServiceLevel.value)
    } catch (error) {
      console.error(error)
      Message.error('请求失败')
    }
  } else {
    // 新增
    if (!currentServiceLevel.value.serviceId) {
      Message.error('请选择服务')
      return
    }
    try {
      resp = await addUserServiceLevel(currentServiceLevel.value)
    } catch (error) {
      console.error(error)
      Message.error('请求失败')
    }
  }
  if (resp && resp.code === 0) {
    Message.success('保存成功')
    getListUserServiceLevel()
  } else {
    console.error(resp?.msg)
    Message.error(resp?.msg || '请求失败')
  }
}

const tagColors = [
  '#00b42a',
  '#165dff',
  '#ff7d00',
  '#f53f3f']

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
            <a-button status="warning" type="outline" @click="showServiceSecretLevelEditor(record)">服务密级配置
            </a-button>
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

  <!-- 服务密级配置弹窗 -->
  <a-drawer :visible="showServiceSecretLevelDrawer" :width="520" unmount-on-close
            @cancel="showServiceSecretLevelDrawer = false">
    <template #title>
      <a-space size="large">
        <span>服务密级配置</span>
        <a-button type="primary" @click="addNewServiceLevel">新增服务密级</a-button>
      </a-space>
    </template>
    <div>
      <a-list size="small">
        <a-list-item v-for="item in userServiceLevelList">
          <div class="flex justify-between">
            <div>
              <div>{{ item.service?.name }}</div>
              <div>{{ item.service?.domain }}</div>
            </div>
            <a-tag :color="tagColors[(item.secLevel || 1) - 1]">{{ item.secLevel }}</a-tag>
          </div>
        </a-list-item>
      </a-list>
    </div>
  </a-drawer>

  <!-- 编辑服务密级 -->
  <a-modal v-model:visible="showServiceLevelModal" title="服务密级" unmount-on-close
           @cancel="showServiceLevelModal = false" @before-ok="saveServiceLevel">
    <a-form :model="currentServiceLevel">
      <a-form-item field="name" label="服务">
        <a-select v-model:model-value="currentServiceLevel.serviceId" :filter-option="false"
                  :field-names="{ label: 'name', value: 'id' }"
                  :loading="serviceLoading" allow-search @search="handleServiceSearch">
          <a-option v-for="item of canSelecteServiceList" :value="item.id">
            <div>
              {{ item.name }}
            </div>
            <div>{{ item.domain }}</div>
          </a-option>
        </a-select>
      </a-form-item>
      <a-form-item field="secLevel" label="密级">
        <a-select v-model="currentServiceLevel.secLevel">
          <a-option :value="1">一级</a-option>
          <a-option :value="2">二级</a-option>
          <a-option :value="3">三级</a-option>
          <a-option :value="4">四级</a-option>
        </a-select>
      </a-form-item>
    </a-form>
  </a-modal>
</template>