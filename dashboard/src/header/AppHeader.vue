<script setup lang="ts">
import { FundFilled, } from '@ant-design/icons-vue';
import AddButton from './AddProductButton.vue';
import LogoutButton from './LogoutButton.vue';
</script>

<template>
    <a-row>
        <a-col :span="12">
            <div class="logo">
                <a-space align="center">
                    <fund-filled :style="{ fontSize: '48px', color: '#fff' }" />
                    <p :style="{ fontSize: '24px', color: '#fff' }">Belle Maison Price Tracker</p>
                </a-space>
            </div>
        </a-col>
        <a-col :span="12">
            <a-row justify="end">
                <a-space>
                    <a-col>
                        <add-button />
                    </a-col>
                    <a-col>
                        <logout-button />
                    </a-col>
                </a-space>
            </a-row>
        </a-col>
    </a-row>
</template>