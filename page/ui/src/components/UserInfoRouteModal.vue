<script lang="ts" setup>
import {Service} from '@/types/service';
import {UserInfoRoute} from '@/types/userInfoRoute';
import {computed, onMounted, ref} from 'vue';
import {get} from 'radash'
import {
  addUserInfoRoute,
  deleteUserInfoRoute,
  getUserInfoRouteByServiceId,
  updateUserInfoRoute
} from '@/api/userInfoRoute';
import {Message} from '@arco-design/web-vue';

const props = defineProps<{
  service: Service | undefined
}>()

const emit = defineEmits(['close'])
const show = ref(false)
const userInfoRoute = ref<UserInfoRoute>({
  id: '',
  serviceId: '',
  path: '',
  usernamePath: '',
  uniKeyPath: '',
  matchKey: '',
  tokenPosition: '',
  method: 'GET'
})
const title = computed(() => (userInfoRoute.value.id ? '编辑' : '新增') + `用户信息路由(${props.service?.name})`)

const cancel = () => {
  emit('close')
}
const submit = async () => {
  let resp = null
  if (userInfoRoute.value.id) {
    // 修改
    try {
      resp = await updateUserInfoRoute(userInfoRoute.value)
    } catch (error) {
      console.log(error)
      Message.error('更新用户信息路由失败')
      return
    }
  } else {
    // 新增
    userInfoRoute.value.serviceId = props.service?.id
    try {
      resp = await addUserInfoRoute(userInfoRoute.value)
    } catch (error) {
      console.log(error)
      Message.error('新增用户信息路由失败')
      return
    }
  }

  if (resp.code === 0) {
    Message.success('操作成功')
    emit('close')
  }
}

const jsonText = ref<string>('')
const isJsonTextValid = computed(() => {
  try {
    JSON.parse(jsonText.value)
    return true
  } catch (error) {
    return false
  }
})

const usernameValue = computed(() => {
  if (userInfoRoute.value.usernamePath && jsonText.value && isJsonTextValid.value) {
    try {
      const json = JSON.parse(jsonText.value)
      return get(json, userInfoRoute.value.usernamePath, '未找到')
    } catch (error) {
      return ''
    }
  }
  return ''
})

const tokenValue = computed(() => {
  if (userInfoRoute.value.tokenPosition && jsonText.value && isJsonTextValid.value) {
    const tpArr = userInfoRoute.value.tokenPosition.split(':')
    if (tpArr.length !== 3) {
      return ''
    }
    if (tpArr[0] === 'response' && tpArr[1] === 'body') {
      try {
        const json = JSON.parse(jsonText.value)
        return get(json, tpArr[2], '未找到')
      } catch (error) {
        return ''
      }
    }
  }
  return ''

})

const deleteRouteInfo = async () => {
  if (!userInfoRoute.value.id) {
    return
  }
  // 删除
  try {
    const resp = await deleteUserInfoRoute(userInfoRoute.value.id)
    if (resp.code === 0) {
      Message.success('删除成功')
      emit('close')
    } else {
      Message.error('删除失败')
    }
  } catch (error) {
    console.log(error)
    Message.error('删除失败')
  }
}

onMounted(async () => {
  if (!props.service || !props.service.id) {
    return
  }
  try {
    const resp = await getUserInfoRouteByServiceId(props.service.id)
    if (resp.code === 0) {
      userInfoRoute.value = resp.data || userInfoRoute.value
      show.value = true
    } else {
      Message.error('获取用户信息路由失败:' + resp.msg)
    }
  } catch (error) {
    console.log(error)
    Message.error('获取用户信息路由失败')
  }
})
</script>

<template>
  <a-modal v-model:visible="show" :title="title" unmount-on-close width="800px" @cancel="cancel">
    <template #footer>
      <a-popconfirm content="确定删除该用户信息路由吗？" @ok="deleteRouteInfo">
        <a-button status="danger" type="outline">删除</a-button>
      </a-popconfirm>
      <a-button @click="cancel">取消</a-button>
      <a-button type="primary" @click="submit">确定</a-button>
    </template>

    <div class="flex items-stretch w-100%">
      <div class="w-200px">
        <span>该路由返回的json结构粘贴此处可实时反馈</span>
        <a-textarea v-model:model-value="jsonText" :auto-size="{ minRows: 10 }" :error="!isJsonTextValid" allow-clear>
        </a-textarea>
      </div>
      <div class="w-100%">
        <a-form :model="userInfoRoute">
          <a-form-item field="path" label="路径" tooltip="用户信息的请求路径">
            <a-input v-model:model-value="userInfoRoute.path"/>
          </a-form-item>
          <a-form-item field="method" label="方法" tooltip="请求方法">
            <a-select v-model:model-value="userInfoRoute.method" style="width: 100%">
              <a-option value="GET">GET</a-option>
              <a-option value="POST">POST</a-option>
              <a-option value="PUT">PUT</a-option>
              <a-option value="DELETE">DELETE</a-option>
            </a-select>
          </a-form-item>
          <a-form-item field="usernamePath" label="用户名路径"
                       tooltip="返回的json中提取用户名数据的路径，如data.user.username">
            <a-input v-model:model-value="userInfoRoute.usernamePath"/>
            <template #extra>{{ usernameValue }}</template>
          </a-form-item>
          <a-form-item field="uniKeyPath" label="唯一标识路径" tooltip="返回的json中提取用户唯一标识，如data.user.id">
            <a-input v-model:model-value="userInfoRoute.uniKeyPath"/>
          </a-form-item>
          <a-form-item field="matchKey" label="匹配键"
                       tooltip="匹配本系统中存储的用户信息字段，-表示直接匹配用户的唯一标识，其他则按路径从用户唯一标识json中匹配">
            <a-input v-model:model-value="userInfoRoute.matchKey"/>
          </a-form-item>
          <a-form-item field="tokenPosition" label="Token位置"
                       tooltip="token获取的位置，如：request:header:Authorization, request:query:token, request:body:auth.token, request:cookies:token, request:cookies:sessionId, response:body:data.token">
            <a-input v-model:model-value="userInfoRoute.tokenPosition"/>
            <template #extra>{{ tokenValue }}</template>
          </a-form-item>
        </a-form>
      </div>
    </div>

  </a-modal>
</template>