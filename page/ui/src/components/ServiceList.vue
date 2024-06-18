<script lang="ts" setup>
import {addService, updateService} from '@/api/service';
import {addField, getFieldList, updateField} from '@/api/serviceField';
import {ServiceField} from '@/types/field';
import {Service} from '@/types/service';
import {Message} from '@arco-design/web-vue';
import {computed, ref} from 'vue';
import {securityLevelToText} from '@/utils/security'

const props = defineProps<{
  selectedService: Service | undefined
  serviceList: Service[]
}>()

const emit = defineEmits(['serviceSelected', 'serviceUpdated'])

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
    emit('serviceUpdated', resp.data)
  } else {
    Message.error('操作失败')
  }
}

// 脱敏
const showDesensitive = ref(false)
const desensitiveTitle = computed(() => {
  return props.selectedService?.name + '的脱敏配置'
})
const desensitiveFieldList = ref<ServiceField[]>([])
const fieldsTotal = ref(0)
const fieldCondition = ref<ServiceField>({
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
  if (!props.selectedService) {
    return
  }
  fieldCondition.value.serviceId = props.selectedService.id
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
const currentField = ref<ServiceField>({})
const showFieldEditModal = ref(false)
const editField = (field: ServiceField) => {
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
</script>
<template>
  <a-list hoverable split>
    <template #header>
      <div class="flex justify-between">
        <span>服务列表</span>
        <a-button size="mini" type="primary" @click="editService({})">添加服务</a-button>
      </div>
    </template>
    <a-list-item v-for="service in serviceList" @click="emit('serviceSelected', service)">
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

  <!-- 添加服务 -->
  <a-modal v-model:visible="showServiceModal" :title="`${currentService.id ? '编辑' : '添加'}服务`" unmount-on-close
           @cancel="showServiceModal = false" @before-ok="handleServiceEditor">
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