<script lang="ts" setup>
import {onMounted, ref} from 'vue';
import {useRoute, useRouter} from 'vue-router';

const route = useRoute();
const router = useRouter();

const changeNav = (key: string) => {
  router.push({name: key});
}

const selectedKeys = ref<string[]>([route.name as string]);

onMounted(() => {
})
</script>

<template>
  <a-layout style="height: 100vh;">
    <a-layout-header>
      <a-menu v-model:selected-keys="selectedKeys" :default-selected-keys="['1']" mode="horizontal"
              @menu-item-click="changeNav">
        <a-menu-item key="0" :style="{ padding: 0, marginRight: '38px' }" disabled>
          安全网关服务
        </a-menu-item>
        <a-menu-item key="Gateway">网关配置</a-menu-item>
        <a-menu-item key="Upstream">上游服务</a-menu-item>
        <a-menu-item key="Certificate">证书管理</a-menu-item>
        <a-menu-item key="User">用户管理</a-menu-item>
      </a-menu>
    </a-layout-header>
    <router-view v-slot="{ Component }">
      <transition>
        <keep-alive>
          <component :is="Component"/>
        </keep-alive>
      </transition>
    </router-view>
  </a-layout>
</template>