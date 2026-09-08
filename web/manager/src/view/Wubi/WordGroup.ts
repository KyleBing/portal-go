import { Word } from '../../types'

// 词条对象 分组
interface Word {
    // Add appropriate properties for Word type
    [key: string]: any;
}

class WordGroup {
    id: string | number;
    groupName: string;
    dict: Word[];

    constructor(id: string | number, groupName: string = '', words: Word[] = []) {
        this.id = id;
        this.groupName = groupName;
        this.dict = words;
    }

    // 复制一个对象
    clone(): WordGroup {
        return new WordGroup(this.id, this.groupName, [...this.dict]);
    }
}

export { WordGroup };
export type { Word }; 