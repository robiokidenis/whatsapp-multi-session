// Import dependencies
import { createApp } from 'vue';
import axios from 'axios';
import QRCode from 'qrcode';

// Make globals available
window.Vue = { createApp };
window.axios = axios;
window.QRCode = QRCode;

console.log('Dependencies loaded:', {
    Vue: !!window.Vue,
    axios: !!window.axios,
    QRCode: !!window.QRCode
});