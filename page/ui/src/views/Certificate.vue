<script lang="ts" setup>
import {addCertificate, deleteCertificate, getCertificateList} from '@/api/certificate';
import {Certificate} from '@/types/certificate';
import {Message, PaginationProps, TableColumnData} from '@arco-design/web-vue';
import {onMounted, ref} from 'vue';
import {Response} from '@/types/common'
import moment from 'moment';

const conditionCollapsed = ref<boolean>(false);
const condition = ref<Certificate>({
  page: 1,
  pageSize: 10,
})
const loading = ref<boolean>(false)
const list = ref<Certificate[]>([])
const pagination = ref<PaginationProps>({
  total: 0,
  pageSize: 10,
})
const columns: TableColumnData[] = [
  {
    title: '证书名称',
    dataIndex: 'certName',
  },
  {
    title: '服务域名',
    dataIndex: 'serveDomain',
  },
  {
    title: '证书描述',
    dataIndex: 'certDesc',
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
    const resp = await getCertificateList(condition.value);
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
const showCertificateModal = ref<boolean>(false);
const currentCertificate = ref<Certificate>({});
const showEditor = (data: Certificate) => {
  currentCertificate.value = data;
  showCertificateModal.value = true;
}
const saveCertificate = async (done: (closed: boolean) => void) => {
  let resp: Response<Certificate> | undefined;
  if (!currentCertificate.value.id) {
    // 新增，检查必填项
    if (!currentCertificate.value.certName || !currentCertificate.value.serveDomain || !currentCertificate.value.certPem || !currentCertificate.value.keyPem) {
      Message.error('请填写完整信息');
      return;
    }
    try {
      resp = await addCertificate(currentCertificate.value);
    } catch (error) {
      console.error(error);
      Message.error('请求失败');
    }
  } else {
    // 编辑
    try {
      resp = await addCertificate(currentCertificate.value);
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
const readyToDelete = async (data: Certificate) => {
  if (!data.id) {
    return;
  }
  try {
    const resp = await deleteCertificate(data.id);
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
          <span class="w-120px text-right">证书名称：</span>
          <a-input v-model="condition.certName" placeholder="证书名称"/>
        </a-grid-item>
        <a-grid-item class="flex items-center">
          <span class="w-120px text-right">服务域名：</span>
          <a-input v-model="condition.serveDomain" placeholder="服务域名"/>
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
  <a-modal v-model:visible="showCertificateModal" title="编辑" unmount-on-close @cancel="showCertificateModal = false"
           @before-ok="saveCertificate">
    <a-form :model="currentCertificate">
      <a-form-item field="certName" label="证书名称">
        <a-input v-model="currentCertificate.certName"/>
      </a-form-item>
      <a-form-item field="serveDomain" label="服务域名">
        <a-input v-model="currentCertificate.serveDomain"/>
      </a-form-item>
      <a-form-item field="certDesc" label="证书描述">
        <a-input v-model="currentCertificate.certDesc"/>
      </a-form-item>
      <a-form-item field="certPem" label="证书内容">
        <a-textarea v-model="currentCertificate.certPem" placeholder="-----BEGIN CERTIFICATE-----"/>
      </a-form-item>
      <a-form-item field="keyPem" label="私钥内容">
        <a-textarea v-model="currentCertificate.keyPem" placeholder="-----BEGIN RSA PRIVATE KEY-----"/>
      </a-form-item>
    </a-form>
  </a-modal>
</template>