// 需要在调用前导入crypto-js（https://lf26-cdn-tos.bytecdntp.com/cdn/expire-1-M/crypto-js/4.1.1/crypto-js.min.js）
// 加密JS文件：https://3d-platform-pro.obs.cn-south-1.myhuaweicloud.com/ecfb29bec27c79ff4fc9f94a20be3e10.min

/**
 * 编码 https://player.ddzyku.com:3653/get_url_v2 接口请求数据
 * 源代码位置：[_0x2f76de(0x165,']^%c')](_0x1ddfea,_0x85c5d,_0x25553c)
 * @param key
 * @param iv
 * @param data
 * @returns {*}
 */
function fuckCryptoEncode(key, iv, data) {
    return CryptoJS.AES.encrypt(data, CryptoJS.enc.Utf8.parse(key), {
        'iv': CryptoJS.enc.Hex.parse(iv),
        'mode': CryptoJS.mode.CBC,
        'padding': CryptoJS.pad.Pkcs7
    }).toString()
}

/**
 * 解码 https://player.ddzyku.com:3653/get_url_v2 接口返回数据
 * 源代码位置：[_0x2f76de(0x3f2,'lA0#')](_0x18e43b,_0x22739d,_0x3780eb)
 * @param key
 * @param iv
 * @param data
 * @returns {*}
 */
function fuckCryptoDecode(key, iv, data) {
    const cipherText = CryptoJS.lib.WordArray.create(CryptoJS.enc.Base64.parse(data)['words']);
    return CryptoJS.AES.decrypt({'ciphertext': cipherText}, CryptoJS.enc.Utf8.parse(key), {
        'iv': CryptoJS.enc.Hex.parse(iv),
        'mode': CryptoJS.mode.CBC,
        'padding': CryptoJS.pad.Pkcs7
    }).toString(CryptoJS.enc.Utf8);

}