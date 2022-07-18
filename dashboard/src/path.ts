let base = import.meta.env.BASE_URL;

if (base.slice(-1) == '/') {
    base = base.slice(0, -1)
};

export const logoutURL = base + '/logout'
export const basePathGetProduct = base + '/api/product/';
export const basePathAddTarget = base + '/api/target/';
export const basePathDeleteTarget = base + '/api/target/';
export const basePathUpdateTarget = base + '/api/target/';
export const basePathGetTargets = base + '/api/targets';
