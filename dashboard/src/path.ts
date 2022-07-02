const base = () => {
    let base = import.meta.env.BASE_URL;
    if (base.slice(-1) == '/') {
        return base.slice(0, -1)
    } else {
        return base
    }
}

export const basePathGetProduct = base + '/api/product/';
export const basePathAddTarget = base + '/api/target/';
export const basePathDeleteTarget = base + '/api/target/';
export const basePathUpdateTarget = base + '/api/target/';
export const basePathGetTargets = base + '/api/targets';
