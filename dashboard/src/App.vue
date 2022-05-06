<script setup lang="ts">
import AppHeader from './header/AppHeader.vue';
import AppMain from './main/AppMain.vue';
import AppFooter from './footer/AppFooter.vue';
import 'ant-design-vue/dist/antd.css';
</script>

<template>
  <a-layout class="layout">
    <a-layout-header>
      <AppHeader />
    </a-layout-header>
    <a-layout-content style="padding: 50px 50px 25px">
      <AppMain />
    </a-layout-content>
    <a-layout-footer style="text-align: center">
      <AppFooter />
    </a-layout-footer>
  </a-layout>
</template>