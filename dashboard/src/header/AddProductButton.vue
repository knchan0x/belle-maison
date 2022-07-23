<script setup lang="ts">
import { reactive } from 'vue';
import { computed } from '@vue/reactivity';
import { message } from 'ant-design-vue';
import ProductInfo, { type StyleOption } from '@/components/ProductInfo.vue';
import { useTargets } from '@/store/targets';
import { fetchProductInfo, urlToProductCode, stylesToOptions, type Response, } from '@/utils/utils';

interface PageState {
  // product
  url: string;
  productCode: string;

  // http response
  data: Response | null;
  errMsg: string;
  isError: boolean;

  // target product info
  options: Array<StyleOption>;
  selectedStyleCode: string;
  targetPrice: number;

  // page control
  warning: string;
  isLoading: boolean;
  isSearching: boolean;
  isVisible: boolean;
  isConfirming: boolean;
}

const pageState = reactive<PageState>({
  // product url
  url: '',
  productCode: '',

  // http response
  data: null,
  errMsg: '',
  isError: false,

  // target product info
  options: [],
  selectedStyleCode: '',
  targetPrice: 0,

  // page control
  warning: '',
  isSearching: false,
  isLoading: true,
  isVisible: false,
  isConfirming: false,
});

const lockAdd = computed(() => {
  if (pageState.selectedStyleCode === '' || pageState.targetPrice === 0) {
    return true;
  }
  return false;
});

const imageUrl = computed(() => {
  if (pageState.selectedStyleCode === '') {
    return '';
  }
  return pageState.data?.Product.Styles.find(
    (e) => e.StyleCode === pageState.selectedStyleCode
  )?.ImageUrl;
});

const price = computed(() => {
  if (pageState.selectedStyleCode === '') {
    return NaN;
  }
  return pageState.data?.Product.Styles.find(
    (e) => e.StyleCode === pageState.selectedStyleCode
  )?.Price;
});

const onSearch = async () => {
  pageState.warning = '';
  pageState.isSearching = true;
  pageState.isLoading = true;
  if (pageState.url !== '') {
    const code = urlToProductCode(pageState.url);
    if (code !== '') {
      pageState.productCode = code;
      await loadProductInfo(code);
      pageState.isLoading = false;
    } else {
      pageState.warning = 'URL incorrect.';
      pageState.isSearching = false;
    }
  }
};

const loadProductInfo = async (productCode: string) => {
  pageState.isSearching = true;
  try {
    pageState.isLoading = true;
    const response = await fetchProductInfo(productCode);
    pageState.data = response;
    pageState.isError = false;
    pageState.errMsg = '';
    pageState.options = stylesToOptions(
      response.Product.Styles.sort(
        (a, b) => parseInt(a.StyleCode) - parseInt(b.StyleCode)
      ) || []
    );
  } catch (e: any) {
    pageState.errMsg = e.message;
  } finally {
    pageState.isLoading = false;
  }
};

const showAddProductPage = () => {
  pageState.isVisible = true;
};

const targets = useTargets();

const handleOk = async () => {
  pageState.isConfirming = true;

  const colour =
    pageState.data?.Product.Styles.find(
      (e) => e.StyleCode === pageState.selectedStyleCode
    )?.Colour || '';
  const size =
    pageState.data?.Product.Styles.find(
      (e) => e.StyleCode === pageState.selectedStyleCode
    )?.Size || '';

  try {
    await targets.add(
      pageState.productCode,
      colour,
      size,
      pageState.targetPrice
    );
    pageState.isConfirming = false;
    resetValues();
    message.success('Successful');
  } catch (e: any) {
    pageState.isConfirming = false;
    if (e.response) {
      let msg = e.response.data as {
        error: String;
      };
      message.error('Error: ' + msg.error + '. Please try again.');
    } else {
      message.error('Error: ' + e.message + '. Please try again.');
    }
  }
};

const handleCancel = () => {
  resetValues();
};

const resetValues = () => {
  pageState.url = '';
  pageState.productCode = '';

  pageState.data = null;
  pageState.errMsg = '';
  pageState.isError = false;

  pageState.options = [];
  pageState.selectedStyleCode = '';
  pageState.targetPrice = 0;

  pageState.warning = '';
  pageState.isSearching = false;
  pageState.isLoading = true;
  pageState.isVisible = false;
  pageState.isConfirming = false;
};
</script>

<template>
  <div>
    <a-button type="primary" @click="showAddProductPage">Add New</a-button>
    <a-modal v-model:visible="pageState.isVisible" title="Add New Product" @ok="handleOk" :closable="false"
      @cancel="handleCancel">
      <template #footer>
        <a-button key="cancel" @click="handleCancel">Cancel</a-button>
        <a-button key="Add" type="primary" :disabled="lockAdd" :loading="pageState.isConfirming" @click="handleOk">Add
        </a-button>
      </template>
      <a-form layout="vertical" :model="pageState">
        <a-form-item label="Product URL" name="url">
          <a-row justify="space-around">
            <a-col :span="20">
              <a-input v-model:value="pageState.url"
                placeholder="https://www.bellemaison.jp/shop/commodity/0000/1097791" :allowClear="true">
              </a-input>
              <p v-if="pageState.warning !== ''" style="color: red">
                {{ pageState.warning }}
              </p>
            </a-col>
            <a-col :span="4">
              <a-button @click="onSearch">Search</a-button>
            </a-col>
          </a-row>
        </a-form-item>
        <div v-if="pageState.isSearching">
          <div v-if="pageState.isLoading" class="loading">
            <a-spin />
          </div>
          <div v-else>
            <div v-if="!pageState.isError">
              <product-info @selected="(selected: string) => pageState.selectedStyleCode = selected[1]"
                @target="(price: number) => pageState.targetPrice = price" :style-options="pageState.options"
                :price="price" :image-url="imageUrl" />
            </div>
            <div v-else>
              <p>{{ pageState.errMsg }}</p>
            </div>
          </div>
        </div>
      </a-form>
    </a-modal>
  </div>
</template>

<style>
.loading {
  text-align: center;
  background: rgba(0, 0, 0, 0.05);
  border-radius: 4px;
  margin-bottom: 20px;
  padding: 30px 50px;
  margin: 20px 0;
}
</style>
