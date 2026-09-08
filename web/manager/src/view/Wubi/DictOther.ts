// 其它字典对象
import { Word } from './Word';

type DictFormat = 'cww' | 'wc' | 'cw' | 'w' | 'rime';

class DictOther {
    dictTypeName: string;
    filePath: string;
    fileName: string;
    lastIndex: number;
    seperator: string;
    dictFormat: DictFormat;
    characterMap: Map<string, string>;
    wordsOrigin: Word[];

    constructor(fileContent: string, fileName: string, filePath: string, seperator: string = ' ', dictFormat: DictFormat = 'rime') {
        this.dictTypeName = 'DictOther';
        this.filePath = filePath; // 文件路径
        this.fileName = fileName; // 文件路径
        this.lastIndex = 0; // 最后一个 Index 的值，用于新添加词时，作为唯一的 id 传入
        this.seperator = seperator; // 默认间隔符为空格
        this.dictFormat = dictFormat; // 码表格式： 一码多词什么的 cww: 一码多词 | wc: 一词一码 | cw: 一码一词
        this.characterMap = new Map(); // 单字码表，用于根据此生成词语码表
        this.wordsOrigin = this.getDictWordsInNormalMode(fileContent);
    }

    // 总的词条数量
    get countDictOrigin(): number {
        return this.wordsOrigin.length;
    }

    // 设置 seperator
    setSeperator(seperator: string): void {
        this.seperator = seperator;
    }

    // 设置 dictFormat
    setDictFormat(dictFormat: DictFormat): void {
        this.dictFormat = dictFormat;
    }

    // 获取指定字数的词条组
    getWordsLengthOf(length: number): Word[] {
        switch (length) {
            case 0:
                return this.wordsOrigin;
            case 1:
            case 2:
            case 3:
            case 4:
                return this.wordsOrigin.filter(word => getUnicodeStringLength(word.word) === length);
            default:
                return this.wordsOrigin.filter(word => getUnicodeStringLength(word.word) > 4);
        }
    }

    // 查重，返回重复定义的字词
    // includeCharacter 当包含单字时
    getRepetitionWords(filterSingleCharacter: boolean, isWithAllRepeatWord: boolean): Word[] {
        let startPoint = new Date().getTime();
        let wordMap = new Map<string, Word>();
        let repetitionWords: Word[] = [];
        
        this.wordsOrigin.forEach(word => {
            if (filterSingleCharacter) {
                if (wordMap.has(word.word) && getUnicodeStringLength(word.word) === 1) {
                    repetitionWords.push(word);
                    if (isWithAllRepeatWord) {
                        let matchedWord = wordMap.get(word.word);
                        if (matchedWord) repetitionWords.push(matchedWord);
                    }
                } else {
                    wordMap.set(word.word, word);
                }
            } else {
                if (wordMap.has(word.word) && getUnicodeStringLength(word.word) > 1) {
                    repetitionWords.push(word);
                    if (isWithAllRepeatWord) {
                        let matchedWord = wordMap.get(word.word);
                        if (matchedWord) repetitionWords.push(matchedWord);
                    }
                } else {
                    wordMap.set(word.word, word);
                }
            }
        });

        repetitionWords.sort((a, b) => {
            return a.toComparableString() > b.toComparableString() ? 1 : -1;
        });
        console.log('重复词条数量:未去重之前 ', repetitionWords.length);

        for (let i = 0; i < repetitionWords.length - 1; i++) {
            if (repetitionWords[i].id === repetitionWords[i + 1].id) {
                repetitionWords.splice(i, 1);
                i = i - 1;
            }
        }
        console.log(`查重完成，用时 ${new Date().getTime() - startPoint} ms`);
        console.log('词条字典数量: ', wordMap.size);
        console.log('重复词条数量: ', repetitionWords.length);
        console.log('重复 + 词条字典 = ', repetitionWords.length + wordMap.size);
        return repetitionWords;
    }

    // 返回所有 word
    getDictWordsInNormalMode(fileContent: string): Word[] {
        let startPoint = new Date().getTime();
        let EOL = this.getFileEOLFrom(fileContent);
        let lines = fileContent.split(EOL); // 拆分词条与编码成单行
        this.lastIndex = lines.length + 1;
        // 如果为纯词模式，就使用所有的行，否则就根据分隔符进行筛选
        let linesValid = this.dictFormat === 'w' ? lines : lines.filter(item => item.indexOf(this.seperator) > -1);
        let words: Word[] = [];
        console.log('正常词条的行数：', linesValid.length);
        
        linesValid.forEach(item => {
            let currentWords = this.getWordsFromLine(item);
            words.push(...currentWords); // 拼接词组
            currentWords.forEach(currentWord => {
                if (getUnicodeStringLength(currentWord.word) === 1
                    && currentWord.code.length >= 2
                    && !this.characterMap.has(currentWord.word)) {
                    this.characterMap.set(currentWord.word, currentWord.code);
                }
            });
        });
        
        console.log(`处理文件完成，共：${words.length} 条，用时 ${new Date().getTime() - startPoint} ms`);
        return words;
    }

    // 排序
    sort(): void {
        let startPoint = new Date().getTime();
        this.wordsOrigin.sort((a, b) => a.code < b.code ? -1 : 1);
        console.log(`排序用时 ${new Date().getTime() - startPoint} ms`);
    }

    // 依次序添加 words
    addWordsInOrder(words: Word[]): void {
        let startPoint = new Date().getTime();
        words.forEach(word => {
            this.addWordToDictInOrder(word);
        });
        console.log(`添加 ${words.length} 条词条到指定码表, 用时 ${new Date().getTime() - startPoint} ms`);
    }

    // 依次序添加 word
    addWordToDictInOrder(word: Word): void {
        let insetPosition: number | null = null; // 插入位置 index
        this.sort(); // 插入之前排序码表
        for (let i = 0; i < this.wordsOrigin.length - 1; i++) { // -1 为了避免下面 i+1 为 undefined
            if (word.code >= this.wordsOrigin[i].code && word.code <= this.wordsOrigin[i + 1].code) {
                insetPosition = i + 1;
                break;
            }
        }
        if (!insetPosition) {  // 没有匹配到任何位置，添加到结尾
            insetPosition = this.wordsOrigin.length;
        }
        let wordInsert = word.clone(); // 断开与别一个 dict 的引用链接，新建一个 word 对象，不然两个 dict 引用同一个 word
        wordInsert.setId(this.lastIndex++); // 给新的 words 一个新的唯一 id
        this.wordsOrigin.splice(insetPosition, 0, wordInsert);
    }

    // 判断码表文件的换行符是 \r\n 还是 \n
    getFileEOLFrom(fileContent: string): string {
        if (fileContent.indexOf('\r\n') > -1) {
            console.log('文件换行符为： \\r\\n');
            return '\r\n';
        } else {
            console.log('文件换行符为： \\n');
            return '\n';
        }
    }

    // 删除词条
    deleteWords(wordIdSet: Set<number>): void {
        this.wordsOrigin = this.wordsOrigin.filter(item => !wordIdSet.has(item.id));
    }

    // 在 origin 中调换两个词条的位置
    exchangePositionInOrigin(word1: Word, word2: Word): void {
        // 确保 word1 在前
        if (parseInt(word1.id.toString()) > parseInt(word2.id.toString())) {
            let temp = word1;
            word1 = word2;
            word2 = temp;
        }
        for (let i = 0; i < this.wordsOrigin.length; i++) {
            let tempWord = this.wordsOrigin[i];
            if (tempWord.isEqualTo(word1)) {
                this.wordsOrigin[i] = word2;
            }
            if (tempWord.isEqualTo(word2)) {
                this.wordsOrigin[i] = word1;
            }
        }
    }

    // 从一条词条字符串中获取 word 对象
    // 一编码对应多词
    getWordsFromLine(lineStr: string): Word[] {
        let wordArray = lineStr.split(this.seperator);
        let words: Word[] = [];
        let code: string, word: string, priority: string, comment: string;
        
        switch (this.dictFormat) {
            case 'cww':
                code = wordArray[0];
                for (let i = 1; i < wordArray.length; i++) {
                    words.push(new Word(this.lastIndex, code, wordArray[i]));
                    this.lastIndex = this.lastIndex + 1;
                }
                return words;
            case 'cw':
                code = wordArray[0];
                word = wordArray[1];
                return [new Word(this.lastIndex++, code, word)];
            case 'wc':
                word = wordArray[0];
                code = wordArray[1];
                return [new Word(this.lastIndex++, code, word)];
            case 'w':
                word = wordArray[0];
                return [new Word(this.lastIndex++, '', word)];
            case 'rime':
                word = wordArray[0];
                code = wordArray[1];
                priority = wordArray[2];
                comment = wordArray[3];
                return [new Word(this.lastIndex++, code, word, priority, comment)];
        }
        return [];
    }
}

// 获取字符串的实际 unicode 长度，如：一个 emoji 表情的正确长度应该为 1
function getUnicodeStringLength(str: string): number {
    let wordLength = 0;
    for (let letter of str) {
        wordLength = wordLength + 1;
    }
    return wordLength;
}

export { DictOther }; 