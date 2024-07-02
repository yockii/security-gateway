<script lang="ts" setup>
import {reactive} from 'vue';
import {ProxyTraceLog} from '@/types/log';
import {Message} from '@arco-design/web-vue';
import {countProxyTraceLog} from '@/api/log';
import moment from 'moment';

const traceLogList = reactive<ProxyTraceLog[]>([]);

const addCard = () => {
  if (traceLogList.length > 0 && traceLogList[traceLogList.length - 1].count === undefined) {
    return
  }
  traceLogList.push({
    id: Math.random(),
  })
}

const removeMe = (id: number) => {
  const index = traceLogList.findIndex((traceLog) => traceLog.id === id);
  traceLogList.splice(index, 1);
}

const countTraceLog = async (traceLog: ProxyTraceLog) => {
  // 必须有时间范围
  if (!traceLog.startTime || !traceLog.endTime) {
    Message.warning('请填写时间范围');
    return
  }
  try {
    const resp = await countProxyTraceLog(traceLog);
    if (resp.code === 0) {
      traceLog.count = resp.data;
      traceLog.inEdit = false;
    } else {
      Message.error(resp.msg);
    }
  } catch (error) {
    Message.error('请求失败');
  }
}
</script>

<template>
  <a-layout-content class="p-16px">
    <div class="flex justify-between">
      <h1>Proxy Trace Log</h1>
      <a-button type="primary" @click="addCard">新增卡片</a-button>
    </div>
    <a-grid :colGap="{ xs: 2, sm: 4, md: 6, lg: 12, xl: 16 }" :cols="{ xs: 1, sm: 1, md: 2, lg: 3, xl: 4 }"
            :rowGap="{ xs: 2, sm: 4, md: 6, lg: 12, xl: 16 }">
      <a-grid-item v-for="traceLog in traceLogList" :key="traceLog.id">
        <a-card>
          <template #title><span class="text-16px font-600">{{
              traceLog.count === undefined ? '未计算，请选择条件' :
                  '满足条件数量共计：' + traceLog.count
            }}</span></template>
          <template #extra>
            <a-button type="primary" @click="removeMe(traceLog.id)">
              <template
                  #icon>
                <icon-delete/>
              </template>
            </a-button>
          </template>
          <div class="flex">
            <div class="flex-1">
              <template v-if="traceLog.inEdit">
                <a-form :model="traceLog">
                  <a-form-item field="customIp" label="用户IP">
                    <a-input v-model="traceLog.customIp"/>
                  </a-form-item>
                  <a-form-item field="domain" label="域名">
                    <a-input v-model="traceLog.domain"/>
                  </a-form-item>
                  <a-form-item field="maskingLevel" label="脱敏等级">
                    <a-select v-model="traceLog.maskingLevel">
                      <a-option :value="0">不指定</a-option>
                      <a-option v-for="idx in 4" :key="idx" :value="idx">{{ idx }}级</a-option>
                    </a-select>
                  </a-form-item>
                  <a-form-item field="path" label="路径">
                    <a-input v-model="traceLog.path"/>
                  </a-form-item>
                  <a-form-item field="port" label="端口">
                    <a-input-number v-model="traceLog.port" :max="65535" :min="1"/>
                  </a-form-item>
                  <a-form-item field="targetUrl" label="目标URL">
                    <a-input v-model="traceLog.targetUrl"/>
                  </a-form-item>
                  <a-form-item field="username" label="用户名">
                    <a-input v-model="traceLog.username"/>
                  </a-form-item>
                  <a-form-item field="startTime" label="时间范围">
                    <a-range-picker :model-value="[traceLog.startTime || 0, traceLog.endTime || 0]" show-time
                                    value-format="timestamp"
                                    @ok="(v) => { traceLog.startTime = (v[0] as number); traceLog.endTime = (v[1] as number) }"/>
                  </a-form-item>
                </a-form>
              </template>
              <template v-else>
                <div v-if="traceLog.customIp">用户IP = {{ traceLog.customIp }}</div>
                <div v-if="traceLog.domain">域名 = {{ traceLog.domain }}</div>
                <div v-if="traceLog.maskingLevel">脱敏等级 = {{ traceLog.maskingLevel }}</div>
                <div v-if="traceLog.path">路径 = {{ traceLog.path }}</div>
                <div v-if="traceLog.port">端口 = {{ traceLog.port }}</div>
                <div v-if="traceLog.targetUrl">目标URL = {{ traceLog.targetUrl }}</div>
                <div v-if="traceLog.username">用户名 = {{ traceLog.username }}</div>
                <div v-if="traceLog.startTime">开始时间 =
                  {{ moment(traceLog.startTime).format('YYYY-MM-DD HH:mm:ss') }}
                </div>
                <div v-if="traceLog.endTime">结束时间 =
                  {{ moment(traceLog.endTime).format('YYYY-MM-DD HH:mm:ss') }}
                </div>
              </template>
            </div>
            <div class="ml-8px flex flex-col">
              <template v-if="traceLog.inEdit">
                <a-button status="success" type="secondary" @click="countTraceLog(traceLog)">
                  <template #icon>
                    <icon-check/>
                  </template>
                </a-button>
                <a-button status="danger" type="secondary" @click="traceLog.inEdit = false">
                  <template #icon>
                    <icon-close/>
                  </template>
                </a-button>
              </template>
              <a-button v-else type="secondary" @click="traceLog.inEdit = true">
                <template #icon>
                  <icon-edit/>
                </template>
              </a-button>
            </div>
          </div>
        </a-card>
      </a-grid-item>
    </a-grid>
  </a-layout-content>
</template>