<script lang="ts" setup>
import {addField, deleteField, getFieldList, updateField} from '@/api/routeField';
import {RouteField} from '@/types/field';
import {Route} from '@/types/route';
import {Message} from '@arco-design/web-vue';
import {computed, nextTick, onMounted, ref} from 'vue';
import {securityLevelToText} from '@/utils/security'

const emit = defineEmits(['close'])

const props = defineProps<{
  route: Route | undefined
}>()

// 脱敏
const showDrawer = ref(false)
const desensitiveTitle = computed(() => {
  return props.route?.uri + '的脱敏配置'
})
const desensitiveFieldList = ref<RouteField[]>([])
const fieldsTotal = ref(0)
const fieldCondition = ref<RouteField>({
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
    desensitiveFieldList.value = resp.data?.items || []
    fieldsTotal.value = resp.data?.total || 0
  } catch (error) {
    console.log(error)
  }
}
// 字段编辑
const currentField = ref<RouteField>({})
const showFieldEditModal = ref(false)
const editField = (field: RouteField) => {
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

    currentField.value.routeId = props.route?.id

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

const close = () => {
  showDrawer.value = false
  emit('close')
}

const delField = async (field: RouteField) => {
  if (!field.id) return
  try {
    const resp = await deleteField(field.id)
    if (resp.code === 0) {
      Message.success('删除成功')
      fieldPageChanegd(fieldCondition.value.page || 1)
    } else {
      Message.error('删除失败')
    }
  } catch (error) {
    console.log(error)
    Message.error('删除失败')
  }
}

onMounted(() => {
  nextTick(async () => {
    if (!props.route) {
      return
    }
    fieldCondition.value.routeId = props.route.id
    try {
      const resp = await getDesensitiveFieldList()
      desensitiveFieldList.value = resp.data?.items || []
      fieldsTotal.value = resp.data?.total || 0
      showDrawer.value = true
    } catch (error) {
      console.log(error)
    }
  })
})
</script>

<template>


  <!-- 脱敏配置抽屉 -->
  <a-drawer :visible="showDrawer" :width="520" @cancel="close">
    <template #title>
      <a-space>
        <span>{{ desensitiveTitle }}</span>
        <a-button size="mini" type="primary" @click="editField({})">添加字段</a-button>
      </a-space>
    </template>
    <template #footer>
      <div class="flex justify-end">
        <a-pagination :current="fieldCondition.page" :page-size="fieldCondition.pageSize" :total="fieldsTotal"
          size="mini" @change="fieldPageChanegd" />
      </div>
    </template>
    <a-space direction="vertical" fill>
      <template v-for="field in desensitiveFieldList">
        <a-card :title="field.fieldName" size="small">
          <template #extra>
            <a-button-group>
              <a-button size="mini" type="primary" @click="editField(field)">编辑</a-button>
              <a-popconfirm title="确认删除该字段吗？" @ok="delField(field)">
                <a-button size="mini" status="danger" type="text">删除</a-button>
              </a-popconfirm>
            </a-button-group>
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
        <a-input v-model:modelValue="currentField.fieldName" :disabled="!!currentField.id" />
      </a-form-item>
      <a-form-item field="comment" label="字段描述">
        <a-input v-model:modelValue="currentField.comment" />
      </a-form-item>
      <a-form-item field="level1" label="一级密级">
        <SecurityLevel v-model:modelValue="currentField.level1" />
      </a-form-item>
      <a-form-item field="level2" label="二级密级">
        <SecurityLevel v-model:modelValue="currentField.level2" />
      </a-form-item>
      <a-form-item field="level3" label="三级密级">
        <SecurityLevel v-model:modelValue="currentField.level3" />
      </a-form-item>
      <a-form-item field="level4" label="四级密级">
        <SecurityLevel v-model:modelValue="currentField.level4" />
      </a-form-item>
    </a-form>
  </a-modal>

</template>