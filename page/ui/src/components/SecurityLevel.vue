<script lang="ts" setup>
import {onMounted, ref} from 'vue';

const emit = defineEmits(['update:modelValue'])
const props = defineProps<{
  modelValue: string
}>()
const t = ref('-')
const v = ref('*')
const c = ref(1)

const emitModelValue = () => {
  const tv = t.value
  if (tv === '-') {
    emit('update:modelValue', tv)
    return
  }
  if (tv === 'all-') {
    emit('update:modelValue', `${tv}${v.value}`)
    return
  }

  // 重复c遍v字符
  let str = ''
  for (let i = 0; i < c.value; i++) {
    str += v.value
  }
  if (tv === 'start-') {
    emit('update:modelValue', `${tv}${str}`)
    return
  }
  if (tv === 'middle-') {
    emit('update:modelValue', `${tv}${str}`)
    return
  }
  if (tv === 'end-') {
    emit('update:modelValue', `${tv}${str}`)
    return
  }
  if (tv === 'each-') {
    emit('update:modelValue', `${tv}${str}`)
    return
  }
  if (tv === 'each-^') {
    emit('update:modelValue', `${tv}${str}`)
    return
  }
}

onMounted(() => {
  const st = props.modelValue
  if (!st) {
    t.value = '-'
    return
  }
  if (st === '-') {
    t.value = st
    return
  }

  const tt = st.split('-')
  t.value = tt[0] + '-'
  let tl = tt[1]
  if (tl.indexOf('^') === 0) {
    t.value += '^'
    tl = tl.substr(1)
  }

  if (t.value === 'all-') {
    v.value = tt[1]
    return
  }
  // 计算tl的字符数量
  c.value = tl.length
  v.value = tl[0]

})
</script>

<template>
  <a-space>
    <a-select v-model:model-value="t" :style="{ minWidth: '200px' }" placeholder="请选择脱敏方式"
              @change="emitModelValue">
      <a-option value="-">不脱敏</a-option>
      <a-option value="all-">全部替换为{{ v }}</a-option>
      <a-option value="start-">起始{{ c }}个字符替换为{{ v }}</a-option>
      <a-option value="middle-">中间{{ c }}个字符替换为{{ v }}</a-option>
      <a-option value="end-">最后{{ c }}个字符替换为{{ v }}</a-option>
      <a-optgroup label="每字符替换">
        <a-option value="each-">每字符替换为{{ c }}个{{ v }}</a-option>
        <a-option value="each-^">每{{ c }}个字符替换为一个{{ v }}</a-option>
      </a-optgroup>
    </a-select>
    <a-input-number v-if="t !== 'all-' && t !== '-'" v-model:modelValue="c" :max="10" :min="1"
                    @change="(v: number | undefined) => { emitModelValue(); c = v || 1 }"/>
    <a-input v-if="t !== '-'" v-model:modelValue="v" :max-length="t === 'all-' ? 10 : 1"
             @change="(vv: string) => { emitModelValue(); v = vv }"/>
  </a-space>
</template>