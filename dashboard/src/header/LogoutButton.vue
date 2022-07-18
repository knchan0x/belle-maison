<script setup lang="ts">
import axios from 'axios';
import { logoutURL } from '@/path';
import { message } from 'ant-design-vue';

const logout = () => {
    axios({
        method: 'POST',
        url: logoutURL,
    }).catch((e: any) => { 
        message.error('Error: ' + e.message + '. Please try again.'); 
    })
};
</script>

<template>
    <div>
        <a-button type="secondary" @click="logout">Logout</a-button>
    </div>
</template>