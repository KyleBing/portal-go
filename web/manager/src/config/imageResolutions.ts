export interface ImageResolution {
    name: string;
    width: number;
    height: number;
    quality: number;
}

export const ImageResolutions = {
    THUMBNAIL_50: {
        name: 'thumbnail_50px',
        width: 50,
        height: 50,
        quality: 75
    },
    THUMBNAIL_200: {
        name: 'thumbnail_200px',
        width: 200,
        height: 200,
        quality: 75
    },
    THUMBNAIL_600: {
        name: 'thumbnail_600px',
        width: 600,
        height: 600,
        quality: 75
    },
    THUMBNAIL_500: {
        name: 'thumbnail_500px',
        width: 500,
        height: 500,
        quality: 75
    },
    THUMBNAIL_1000: {
        name: 'thumbnail_1000px',
        width: 1000,
        height: 1000,
        quality: 75
    },
    THUMBNAIL_1500: {
        name: 'thumbnail_1500px',
        width: 1500,
        height: 1500,
        quality: 75
    },
    THUMBNAIL_2000: {
        name: 'thumbnail_2000px',
        width: 2000,
        height: 2000,
        quality: 75
    }
} as const;

export type ImageResolutionKey = keyof typeof ImageResolutions; 