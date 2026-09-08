const BucketQiniuMap = new Map<string, BucketInfo>([
    ['diary-container', {
        baseUrl: 'http://diary-container.kylebing.cn', // 空间域名，最后面不带 `/`
        bucketName: 'diary-container', // 七牛云对象存储空间的名称
        suffiex_thumbnail: 'thumbnail_50px',
        suffiex: 'thumbnail_1000px',
    }],
    ['apple-image', {
        baseUrl: 'http://apple-image.kylebing.cn', // 空间域名，最后面不带 `/`
        bucketName: 'apple-image', // 七牛云对象存储空间的名称
        suffiex_thumbnail: 'thumbnail_50px',
        suffiex: 'thumbnail_1000px',
    }]
])


interface BucketInfo {
    bucketName: string;
    baseUrl: string;
    suffiex_thumbnail: string;
}

export {
    BucketQiniuMap
}
