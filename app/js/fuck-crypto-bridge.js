/**
 * 需要在调用前导入crypto-js（https://lf26-cdn-tos.bytecdntp.com/cdn/expire-1-M/crypto-js/4.1.1/crypto-js.min.js）
 * @param key
 * @param iv
 * @param data
 * @returns {*}
 */
function fuckCrypto(key, iv, data) {
    return CryptoJS.AES.encrypt(data, CryptoJS.enc.Utf8.parse(key), {
        'iv': CryptoJS.enc.Hex.parse(iv),
        'mode': CryptoJS.mode.CBC,
        'padding': CryptoJS.pad.Pkcs7
    }).toString()
}