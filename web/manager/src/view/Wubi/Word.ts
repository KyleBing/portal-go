// 词条对象
class Word {
    id: number;
    code: string;
    word: string;
    priority: string;
    comment: string;

    /**
     *
     * @param id Number ID
     * @param code String 编码
     * @param word String 词条
     * @param priority String 权重
     * @param comment String 备注
     */
    constructor(id: number, code: string, word: string, priority: string = '', comment: string = '') {
        this.id = id;
        this.code = code;
        this.word = word;
        this.priority = priority;
        this.comment = comment;
    }

    toComparableString(): string {
        return this.word + '\t' + this.code + '\t' + this.id + '\t' + this.priority + '\t' + this.comment;
    }

    toString(): string {
        return this.id + '\t' + this.word + '\t' + this.code + '\t' + this.priority + '\t' + this.comment;
    }

    toYamlString(): string {
        if (this.priority && this.comment) {
            return this.word + '\t' + this.code + '\t' + this.priority + '\t' + this.comment;
        } else if (this.priority) {
            return this.word + '\t' + this.code + '\t' + this.priority;
        } else if (this.comment) {
            return this.word + '\t' + this.code + '\t' + this.priority + '\t' + this.comment;
        } else {
            return this.word + '\t' + this.code;
        }
    }

    toFileString(seperator: string, codeFirst: boolean): string {
        if (codeFirst) {
            return this.code + seperator + this.word;
        } else {
            return this.word + seperator + this.code;
        }
    }

    setCode(code: string): void {
        this.code = code;
    }

    setId(id: number): void {
        this.id = id;
    }

    // 复制一个对象
    clone(): Word {
        return new Word(this.id, this.code, this.word, this.priority, this.comment);
    }

    isEqualTo(word: Word): boolean {
        return this.id === word.id;
    }

    // compare a word to another word
    isContentEqualTo(word: Word): boolean {
        return this.word === word.word && this.code === word.code;
    }
}

export { Word }; 