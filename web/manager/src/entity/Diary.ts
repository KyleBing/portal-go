export interface DiaryGeneralInfo {
    id:  number,
    date: string, // "2025-07-23T06:52:55.000Z",
    title: string, // "apple-chip-a",
    content: string,
    temperature: number, // -273,
    temperature_outside: number, // 25,
    weather: string, // "sprinkle",
    category: string, //"memo",
    date_create: string, // "2025-07-23T06:53:05.000Z",
    date_modify: string, // "2025-07-23T06:53:05.000Z",
    uid: number,
    is_public: number,
    is_markdown: number
}