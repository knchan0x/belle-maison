import { defineStore } from 'pinia';
import axios from 'axios';
import { basePathGetTargets, basePathDeleteTarget, basePathAddTarget } from '@/config'

export const useTargets = defineStore('Targets', {
    state: () => ({
        list: [] as Array<Product>,
    }),
    actions: {
        async refresh() {
            axios({
                method: 'GET',
                url: basePathGetTargets,
            }).then((resp) => {
                this.list = resp.data as Product[];
            })
        },
        async delete(id: number): Promise<void> {
            return axios({
                method: 'DELETE',
                url: basePathDeleteTarget + id,
            }).then(() => {
                this.list = this.list.filter(item => item.ID !== id);
            }).catch((e: any) => { throw e })
        },
        async add(productCode: string, colour: string, size: string, price: number) {
            const bodyFormData = new FormData();
            bodyFormData.append('colour', colour);
            bodyFormData.append('size', size);
            bodyFormData.append('price', price.toString());

            return axios({
                method: 'POST',
                url: basePathAddTarget + productCode,
                data: bodyFormData,
                headers: { "Content-Type": "multipart/form-data" },
            }).then(() => {
                this.refresh()
            }).catch((e: any) => { throw e })
        },
    },
})

export interface Product {
    ID: number
    ProductCode: number
    Name: string
    Colour: string
    Size: string
    ImageUrl: string
    TargetPrice: number
    Price: number
    Stock: number
}