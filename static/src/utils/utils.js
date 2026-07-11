// 检查是否为有效链接
const isValidUrl = (url) => {
    const urlRegex = /http[s]?:\/\/[^\s]+/;
    return urlRegex.test(url);
};

// 检查是否为支持的平台
const isSupportedPlatform = (url) => {
    const supportedPlatformsRegex = /(163)|(qq)|(qishui)|(douyin)|(kugou)|(kuwo)|(migu)|(joox)|(5sing)|(91q)|(jamendo)|(apple)|(bilibili)|(b23)/;
    return supportedPlatformsRegex.test(url);
};


export {isValidUrl, isSupportedPlatform};