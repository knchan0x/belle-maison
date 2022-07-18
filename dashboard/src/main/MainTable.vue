<script setup lang="ts">
import { onBeforeMount } from 'vue';
import { message } from 'ant-design-vue';
import { useTargets, type Product } from '../store/targets';

const columns = [
  {
    title: 'Product ID',
    dataIndex: 'ProductCode',
  },
  {
    title: 'Image',
    dataIndex: 'ImageUrl',
  },
  {
    title: 'Product',
    dataIndex: 'Product',
    width: '40%'
  },
  {
    title: 'Current Price',
    dataIndex: 'Price',
  },
  {
    title: 'Target Price',
    dataIndex: 'TargetPrice',
  },
  {
    title: 'Actions',
    dataIndex: 'Actions',
  }
];

const targets = useTargets();

const confirmDelete = async (id: number) => {
  try {
    await targets.delete(id);
    message.success('Successful');
  } catch (e: any) {
    message.error('Error: ' + e.message + '. Please try again.');
  }
}

const productCodeToURL = (code: string) => {
  return "https://www.bellemaison.jp/shop/commodity/0000/" + code
}

onBeforeMount(() => {
  targets.refresh();
})
</script>

<template>
  <a-table :columns="columns" :row-key="(record: Product) => record.ID" :data-source="targets.list">
    <template #bodyCell="{ text, column, index, record }">
      <template v-if="column.dataIndex === 'ProductCode'">
        <a :href="productCodeToURL(text)" target="_blank">{{ text }}</a>
      </template>
      <template v-if="column.dataIndex === 'ImageUrl'">
        <a-image :width="100" :src="text" />
      </template>
      <template v-if="column.dataIndex === 'Product'">
        <p>{{ record.Name }}</p>
        <a-row>
          <a-col>
            <p>Colour: {{ record.Colour }}</p>
          </a-col>
          <a-col>
            <a-divider type="vertical" />
          </a-col>
          <a-col>
            <p>Size: {{ record.Size }}</p>
          </a-col>
        </a-row>
      </template>
      <template v-if="column.dataIndex === 'Price'">
        <p>{{ Intl.NumberFormat('ja-JP', { style: 'currency', currency: 'JPY' }).format(record.Price) }}</p>
      </template>
      <template v-if="column.dataIndex === 'TargetPrice'">
        <p>{{ Intl.NumberFormat('ja-JP', { style: 'currency', currency: 'JPY' }).format(record.TargetPrice) }}</p>
      </template>
      <template v-if="column.dataIndex === 'Actions'">
        <span>
          <a-popconfirm title="Confirmed?" ok-text="Yes" cancel-text="No" @confirm="confirmDelete(record.ID)">
            <a>Delete</a>
          </a-popconfirm>
        </span>
      </template>
    </template>
  </a-table>
</template>