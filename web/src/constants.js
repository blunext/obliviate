// Constants.js
const prod = {
    API_URL: '/variables',
};
const dev = {
    API_URL: 'http://localhost:3000/variables'
};
export const config = process.env.NODE_ENV === 'development' ? dev : prod;
