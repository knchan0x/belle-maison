import { basePathGetProduct } from '@/config';
import type { StyleOption } from '../components/ProductInfo.vue';
import axios from 'axios';

export interface Response {
    ProductCode: string
    Product: {
        Name: string
        Styles: Style[]
    }
    Err: string
};

interface Style {
    StyleCode: string
    ImageUrl: string
    Colour: string
    Size: string
    Price: number
    Stock: number
};

export const fetchProductInfo = async (productCode: string) => {
    try {
        const response = await axios({
            method: 'GET',
            url: basePathGetProduct + productCode,
        });
        return response.data as Response;
    } catch (e: any) {
        if (e.response.status === 404 || e.response.status === 500) {
            const temp = e.response.data as {
                error: string
            };
            throw new Error(temp.error);
        } else {
            throw e;
        }
    }
};

export const stylesToOptions = (styles: Array<Style>) => {
    if (styles.length > 0) {
        const result = styles.reduce((accum, curr) => {
            accum[curr.Colour] = accum[curr.Colour] || [];
            accum[curr.Colour].push(curr);
            return accum
        }, Object.create(null));

        const options: Array<StyleOption> = [];
        if (result !== null) {
            const keys = Object.keys(result);

            keys.forEach(key => {
                const tempChildren: Array<StyleOption> = []
                result[key].forEach((style: Style) => {
                    const tempChild: StyleOption = {
                        value: style.StyleCode,
                        label: style.Size,
                    }
                    tempChildren.push(tempChild);
                });

                const temp: StyleOption = {
                    value: key,
                    label: key,
                    children: tempChildren
                };

                options.push(temp);
            })
        }
        return options
    } else {
        return []
    }
};

const isBelleUrl = (url: string) => {
    if (url.startsWith('https://www.bellemaison.jp/shop/commodity/0000') ||
        url.startsWith('http://www.bellemaison.jp/shop/commodity/0000') ||
        url.startsWith('www.bellemaison.jp/shop/commodity/0000') ||
        url.startsWith('bellemaison.jp/shop/commodity/0000')
    ) {
        return true
    }
    return false
}

export const urlToProductCode = (url: string) => {
    if (url === '' || !isBelleUrl(url)) {
        return '';
    }

    const temp = url.split('/?');
    if (temp.length < 1) {
        return '';
    }

    const parts = temp[0].split('/');
    let offset = 1;
    if (parts[parts.length - 1] === '') {
        offset += 1;
    }

    if (parts[parts.length - offset].length !== 7) {
        return '';
    }

    return parts[parts.length - offset];
};