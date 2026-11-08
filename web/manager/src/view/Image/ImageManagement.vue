<template>
    <Container>
        <Toolbar>
            <template #left>
                <ElForm inline class="tool-bar" @submit.prevent="getFileList">
                    <ElFormItem label="文件描述">
                        <ElInput type="text" v-model="searchKeyword"></ElInput>
                    </ElFormItem>
                    <ElFormItem label="">
                        <ElButton icon="Search" type="primary" @click="getFileList">搜索</ElButton>
                    </ElFormItem>
                </ElForm>
            </template>
            <template #center>
            </template>
            <template #right>
                <ElButton icon="Plus" type="success" @click="addNewImage">添加</ElButton>
                <ElButton
                    icon="Close"
                    :disabled="selectedRows.length === 0"
                    @click="clearSelection"
                >清选</ElButton>
                <ElButton
                    icon="Delete"
                    type="danger"
                    :disabled="selectedRows.length === 0"
                    @click="goBatchDelete"
                >删除</ElButton>
                <ElDropdown trigger="hover" class="ml-2 mr-2">
                    <ElButton
                        icon="CopyDocument"
                        type="primary"
                        :disabled="selectedRows.length === 0">
                        复制所选
                        <ElIcon class="el-icon--right"><ArrowDown /></ElIcon>
                    </ElButton>
                    <template #dropdown>
                        <ElDropdownMenu>
                            <ElDropdownItem
                                v-for="copyOption in copyOptions"
                                :key="`batch-${copyOption.key}`"
                                @click="copyText(getBatchClipboardText(copyOption))">
                                {{ copyOption.label }}
                            </ElDropdownItem>
                        </ElDropdownMenu>
                    </template>
                </ElDropdown>
                <ElCheckbox v-model="copyAsMarkdown">复制为 Markdown 图片</ElCheckbox>
            </template>
        </Toolbar>

        <Content padding="0" :is-show-pagination="true">
            <div class="image-grid" v-loading="isLoading">
                <div v-for="item in tableData" 
                     :key="item.id" 
                     class="image-card"
                     :class="{ 'selected': selectedRows.some(row => row.id === item.id) }"
                     @mousedown="onCardMouseDown"
                     @click="toggleSelection(item, $event)">
                    <div v-if="getSelectionOrder(item.id) > 0" class="selection-order-badge">
                        {{ getSelectionOrder(item.id) }}
                    </div>
                    <div class="image-preview">
                        <img :src="getSafeImageUrl(item.id, ImageResolutions.THUMBNAIL_200.name)"
                             :alt="item.description || 'image'">
                    </div>
                    <div class="image-info">
                        <div class="image-description">{{ item.description || '无描述' }}</div>
                        <div class="image-date">{{ dateFormatter(new Date(item.date_create)) }}</div>
                        <div class="image-actions">
                            <ElButton size="small" @click.stop="openEditDialog(item)" text type="primary" icon="Edit" plain></ElButton>
                            <ElButton size="small" @click.stop="goDelete([item])" text type="danger" icon="delete" plain></ElButton>
                            <div class="copy-split" @click.stop>
                                <ElButton
                                    size="small"
                                    type="info"
                                    icon="CopyDocument"
                                    text
                                    plain
                                    class="copy-split__main"
                                    title="复制图片标识"
                                    @click="copyText(item.id)"
                                />
                                <ElDropdown trigger="click" placement="bottom-end" popper-class="copy-dropdown-compact">
                                    <ElButton size="small" type="info" text plain class="copy-split__caret">
                                        <ElIcon><ArrowDown /></ElIcon>
                                    </ElButton>
                                    <template #dropdown>
                                        <ElDropdownMenu>
                                            <ElDropdownItem
                                                v-for="(resolution, key) in ImageResolutions"
                                                :key="key"
                                                @click="copyText(`${currentBucketInfo?.baseUrl}/${item.id}-${resolution.name}`)"
                                            >
                                                {{ resolution.name }}
                                            </ElDropdownItem>
                                            <ElDropdownItem
                                                @click="copyText(`${currentBucketInfo?.baseUrl}/${item.id}`)"
                                            >
                                                原图
                                            </ElDropdownItem>
                                        </ElDropdownMenu>
                                    </template>
                                </ElDropdown>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </Content>

         <!--  PAGINATION  -->
         <FooterPagination
                :pager-option="pager"
                @size-change="pageSizeChange"
                @pager-change="pageNoChange"
            />

        <!-- Edit Dialog -->
        <ElDialog
            title="编辑图片描述"
            v-model="isShowingEditDialog"
            width="600px"
            center>
            <div class="edit-dialog-content">
                <div class="edit-preview">
                    <img :src="getSafeImageUrl(editForm.id, ImageResolutions.THUMBNAIL_200.name)" alt="preview">
                </div>
                <div class="edit-form-container">
                    <ElForm label-width="80px">
                        <ElFormItem label="图片描述">
                            <ElInput type="text" v-model="editForm.description"></ElInput>
                        </ElFormItem>
                    </ElForm>
                </div>
            </div>
            <template #footer>
                <ElButton @click="isShowingEditDialog = false">取 消</ElButton>
                <ElButton type="primary" @click="submitEdit">确 定</ElButton>
            </template>
        </ElDialog>

        <ElDialog
            title="上传图片"
            v-model="isShowingModalUpload"
            width="760px"
            :close-on-click-modal="!isUploading"
            :close-on-press-escape="!isUploading"
            @closed="resetUploadForm"
            center>
            <div class="upload-dialog">
                <div
                    class="upload-dropzone"
                    :class="{ 'is-dragover': isDragOver, 'is-disabled': isUploading }"
                    @click="triggerFileSelect"
                    @dragenter.prevent="onDragEnter"
                    @dragover.prevent="onDragOver"
                    @dragleave.prevent="onDragLeave"
                    @drop.prevent="onDropFiles"
                >
                    <input
                        type="file"
                        ref="inputUpload"
                        class="upload-dropzone__input"
                        multiple
                        accept="image/*"
                        :disabled="isUploading"
                        @change="inputFileChange"
                    />
                    <ElIcon class="upload-dropzone__icon" :size="36"><Upload /></ElIcon>
                    <p class="upload-dropzone__title">点击或拖拽图片到此处</p>
                    <p class="upload-dropzone__hint">支持多选 / 继续追加 · 单张不超过 10MB · 仅图片</p>
                </div>

                <div v-if="pendingFiles.length" class="upload-meta">
                    <span>已选 {{ pendingFiles.length }} 张</span>
                    <ElButton
                        text
                        type="danger"
                        size="small"
                        :disabled="isUploading"
                        @click="clearPendingFiles"
                    >清空</ElButton>
                </div>

                <div v-if="pendingFiles.length" class="upload-file-grid">
                    <div
                        v-for="item in pendingFiles"
                        :key="item.uid"
                        class="upload-file-card"
                        :class="`is-${item.status}`"
                    >
                        <div class="upload-file-card__thumb">
                            <img :src="item.preview" :alt="item.name">
                            <button
                                v-if="!isUploading"
                                type="button"
                                class="upload-file-card__remove"
                                title="移除"
                                @click.stop="removePendingFile(item.uid)"
                            >
                                <ElIcon :size="12"><Close /></ElIcon>
                            </button>
                            <div v-if="item.status === 'uploading'" class="upload-file-card__mask">上传中</div>
                            <div v-else-if="item.status === 'success'" class="upload-file-card__mask is-success">完成</div>
                            <div v-else-if="item.status === 'error'" class="upload-file-card__mask is-error">失败</div>
                        </div>
                        <ElInput
                            v-model="item.description"
                            size="small"
                            :disabled="isUploading"
                            :title="item.name"
                            placeholder="图片描述"
                        />
                    </div>
                </div>

                <ElProgress
                    v-if="isUploading"
                    :percentage="uploadProgress"
                    :stroke-width="10"
                    striped
                    striped-flow
                />
            </div>
            <template #footer>
                <ElButton :disabled="isUploading" @click="isShowingModalUpload = false">取 消</ElButton>
                <ElButton
                    icon="Upload"
                    type="primary"
                    :loading="isUploading"
                    :disabled="pendingFiles.length === 0"
                    @click="fileUpload"
                >{{ isUploading ? `上传中 ${uploadDoneCount}/${uploadTotalCount}` : '开始上传' }}</ElButton>
            </template>
        </ElDialog>

    </Container>
</template>

<script setup lang="ts">
import {ref, computed, onMounted, onUnmounted, h} from 'vue';
import { useProjectStore } from "@/pinia";
import imageApi, { ImageQiniuRequest } from "@/api/imageQiniuApi.ts";
import * as qiniu from 'qiniu-js';
import Container from "@/layout/Container.vue";
import Toolbar from "@/layout/Toolbar.vue";
import FooterPagination from "@/layout/FooterPagination.vue";
import {BucketQiniuMap} from "@/configProject.ts";
import Content from "@/layout/Content.vue";
import imageQiniuApi from "@/api/imageQiniuApi.ts";
import { ImageResolutions, } from '@/config/imageResolutions';
import { ArrowDown, Close, Upload } from '@element-plus/icons-vue'
import {ElMessage, ElMessageBox, ElNotification} from "element-plus";
import {dateFormatter} from "@/utility.ts";

type UploadItemStatus = 'ready' | 'uploading' | 'success' | 'error';

interface PendingUploadFile {
    uid: string;
    file: File;
    name: string;
    description: string;
    preview: string;
    status: UploadItemStatus;
}

interface TableData {
    id: string;
    description: string;
    type: string;
    date_create: string;
}

const UPLOAD_CONCURRENCY = 3;
const MAX_IMAGE_SIZE = 1024 * 1024 * 10;
const DEFAULT_BUCKET = 'diary-container';

const projectStore = useProjectStore();
const isLoading = ref(false);
const tableData = ref<TableData[]>([]);
const isShowingModalUpload = ref(false);
const searchKeyword = ref('');
const inputUpload = ref<HTMLInputElement | null>(null);
const pendingFiles = ref<PendingUploadFile[]>([]);
const isDragOver = ref(false);
const isUploading = ref(false);
const uploadDoneCount = ref(0);
const uploadTotalCount = ref(0);
const uploadProgress = computed(() => {
    if (!uploadTotalCount.value) return 0;
    return Math.min(100, Math.round((uploadDoneCount.value / uploadTotalCount.value) * 100));
});

const currentBucketInfo = computed(() => {
    return BucketQiniuMap.get(DEFAULT_BUCKET) || BucketQiniuMap.values().next().value;
})

function getSafeImageUrl(id: string, resolution: string) {
    return currentBucketInfo.value ? `${currentBucketInfo.value.baseUrl}/${id}-${resolution}` : '';
}

function getSafeBucketName() {
    return currentBucketInfo.value?.bucketName || DEFAULT_BUCKET;
}

const pager = ref({
    total: 0,
    pageNo: 1,
    pageSize: 30
});

function pageSizeChange(pageSize: number){
    pager.value.pageSize = pageSize
    getFileList()
}

function pageNoChange(pageNo: number){
    pager.value.pageNo = pageNo
    getFileList()
}

const selectedRows = ref<TableData[]>([]);
const lastAnchorIndex = ref(-1);
const copyAsMarkdown = ref(false);

const copyOptions = computed(() => {
    const resolutionOptions = Object.values(ImageResolutions).map(resolution => ({
        key: resolution.name,
        label: `${resolution.name} (${resolution.width}x${resolution.height})`,
        mode: 'resolution' as const,
        resolutionName: resolution.name
    }));

    return [
        ...resolutionOptions,
        {
            key: 'origin',
            label: '原图',
            mode: 'origin' as const
        },
        {
            key: 'id',
            label: '图片标识（无前缀）',
            mode: 'id' as const
        }
    ];
});

function getImageCopyPath(item: TableData, option: { mode: 'resolution' | 'origin' | 'id'; resolutionName?: string }) {
    if (option.mode === 'id') {
        return item.id;
    }
    const baseUrl = currentBucketInfo.value?.baseUrl || '';
    if (option.mode === 'origin') {
        return `${baseUrl}/${item.id}`;
    }
    return `${baseUrl}/${item.id}-${option.resolutionName}`;
}

function getBatchClipboardText(option: { mode: 'resolution' | 'origin' | 'id'; resolutionName?: string }) {
    return selectedRows.value
        .map(item => {
            const path = getImageCopyPath(item, option);
            if (copyAsMarkdown.value && option.mode !== 'id') {
                const imageName = item.description?.trim() || item.id;
                return `![${imageName}](${path})`;
            }
            return path;
        })
        .join('\n');
}

async function copyText(text: string) {
    const value = String(text || '').trim();
    if (!value) {
        ElMessage.warning('没有可复制的内容');
        return;
    }
    try {
        await navigator.clipboard.writeText(value);
        ElMessage.success('复制成功');
    } catch {
        ElMessage.error('复制失败，请手动复制');
    }
}

function onCardMouseDown(event: MouseEvent) {
    // Shift 框选时避免连带选中文字
    if (event.shiftKey) event.preventDefault();
}

function clearSelection() {
    selectedRows.value = [];
    lastAnchorIndex.value = -1;
}

const toggleSelection = (item: TableData, event?: MouseEvent) => {
    const currentIndex = tableData.value.findIndex(row => row.id === item.id);
    if (currentIndex < 0) return;

    if (event?.shiftKey && lastAnchorIndex.value >= 0) {
        const start = Math.min(lastAnchorIndex.value, currentIndex);
        const end = Math.max(lastAnchorIndex.value, currentIndex);
        selectedRows.value = tableData.value.slice(start, end + 1);
        return;
    }

    lastAnchorIndex.value = currentIndex;
    const index = selectedRows.value.findIndex(row => row.id === item.id);
    if (index === -1) {
        selectedRows.value.push(item);
    } else {
        selectedRows.value.splice(index, 1);
    }
};

const getSelectionOrder = (id: string) => {
    const index = selectedRows.value.findIndex(row => row.id === id);
    return index >= 0 ? index + 1 : 0;
};

const getFileList = () => {
    isLoading.value = true;
    const keywordList = searchKeyword.value
        .trim()
        .split(/\s+/)
        .filter(Boolean);
    const params = {
        pageNo: pager.value.pageNo,
        pageSize: pager.value.pageSize,
        keywords: JSON.stringify(keywordList),
        bucket: getSafeBucketName()
    };
    imageApi
        .list(params)
        .then((res: any) => {
            isLoading.value = false;
            tableData.value = res.data.list;
            pager.value.total = res.data.pager.total;
            selectedRows.value = [];
            lastAnchorIndex.value = -1;
        })
        .catch(() => {
            isLoading.value = false;
        });
};

function createPendingUid() {
    return `${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;
}

function defaultDescriptionFromName(fileName: string) {
    const withoutExt = fileName.split('.').slice(0, -1).join('.');
    return withoutExt || fileName;
}

function revokePendingPreviews(files: PendingUploadFile[] = pendingFiles.value) {
    files.forEach(item => URL.revokeObjectURL(item.preview));
}

function resetUploadForm() {
    revokePendingPreviews();
    pendingFiles.value = [];
    isDragOver.value = false;
    isUploading.value = false;
    uploadDoneCount.value = 0;
    uploadTotalCount.value = 0;
    if (inputUpload.value) {
        inputUpload.value.value = '';
    }
}

function clearPendingFiles() {
    if (isUploading.value) return;
    revokePendingPreviews();
    pendingFiles.value = [];
    if (inputUpload.value) {
        inputUpload.value.value = '';
    }
}

function removePendingFile(uid: string) {
    if (isUploading.value) return;
    const index = pendingFiles.value.findIndex(item => item.uid === uid);
    if (index < 0) return;
    URL.revokeObjectURL(pendingFiles.value[index].preview);
    pendingFiles.value.splice(index, 1);
}

function appendSelectedFiles(fileList: File[]) {
    const nextItems: PendingUploadFile[] = [];
    for (const file of fileList) {
        if (!/image\/.*/.test(file.type)) {
            ElMessage.warning(`文件 ${file.name} 不是图片文件`);
            continue;
        }
        if (file.size > MAX_IMAGE_SIZE) {
            ElMessage.warning(`文件 ${file.name} 大小超过 10MB`);
            continue;
        }
        nextItems.push({
            uid: createPendingUid(),
            file,
            name: file.name,
            description: defaultDescriptionFromName(file.name),
            preview: URL.createObjectURL(file),
            status: 'ready',
        });
    }
    if (nextItems.length) {
        pendingFiles.value.push(...nextItems);
    }
}

function triggerFileSelect() {
    if (isUploading.value) return;
    inputUpload.value?.click();
}

function onDragEnter() {
    if (!isUploading.value) isDragOver.value = true;
}

function onDragOver() {
    if (!isUploading.value) isDragOver.value = true;
}

function onDragLeave() {
    isDragOver.value = false;
}

function onDropFiles(event: DragEvent) {
    isDragOver.value = false;
    if (isUploading.value) return;
    const files = Array.from(event.dataTransfer?.files || []);
    if (files.length) appendSelectedFiles(files);
}

const addNewImage = () => {
    resetUploadForm();
    isShowingModalUpload.value = true;
};

function inputFileChange(event: Event) {
    const input = event.target as HTMLInputElement;
    if (!input.files?.length) return;
    appendSelectedFiles(Array.from(input.files));
    input.value = '';
}

function qiniuUploadFile(file: File, token: string): Promise<string> {
    return new Promise((resolve, reject) => {
        qiniu.upload(file, null, token, {}, {}).subscribe({
            next: () => {},
            error: (err: unknown) => reject(err),
            complete: (res: { key: string }) => resolve(res.key),
        });
    });
}

async function registerUploadedImage(content: ImageQiniuRequest) {
    await imageQiniuApi.add(content);
}

async function uploadSinglePendingFile(item: PendingUploadFile, token: string, bucket: string) {
    item.status = 'uploading';
    try {
        const key = await qiniuUploadFile(item.file, token);
        await registerUploadedImage({
            id: key,
            type: 'image',
            bucket,
            description: item.description,
        } as ImageQiniuRequest);
        item.status = 'success';
        return null;
    } catch (err) {
        console.error('Upload error for file:', item.name, err);
        item.status = 'error';
        return {
            fileName: item.name,
            description: item.description || '',
        };
    } finally {
        uploadDoneCount.value += 1;
    }
}

async function runWithConcurrency<T>(items: T[], limit: number, worker: (item: T) => Promise<void>) {
    let cursor = 0;
    const runners = Array.from({ length: Math.min(limit, items.length) }, async () => {
        while (cursor < items.length) {
            const current = items[cursor++];
            await worker(current);
        }
    });
    await Promise.all(runners);
}

// 上传文件：同一 bucket 复用 token，有限并发上传，整批结束后再关窗刷新
async function fileUpload() {
    if (pendingFiles.value.length === 0) {
        ElMessage.warning('请先选择文件');
        return;
    }
    if (isUploading.value) return;

    const bucket = getSafeBucketName();
    const queue = pendingFiles.value.filter(item => item.status !== 'success');
    if (!queue.length) {
        ElMessage.info('没有待上传的文件');
        return;
    }

    queue.forEach(item => {
        item.status = 'ready';
    });
    isUploading.value = true;
    uploadDoneCount.value = 0;
    uploadTotalCount.value = queue.length;

    const failedUploads: { fileName: string; description: string }[] = [];

    try {
        const tokenRes: any = await imageQiniuApi.getUploadToken({
            bucket,
            hahaha: true,
        });
        const token = tokenRes.data;
        if (!token) {
            throw new Error('empty upload token');
        }

        await runWithConcurrency(queue, UPLOAD_CONCURRENCY, async (item) => {
            const failed = await uploadSinglePendingFile(item, token, bucket);
            if (failed) failedUploads.push(failed);
        });

        if (failedUploads.length > 0) {
            ElNotification({
                title: `${failedUploads.length} 个文件上传失败`,
                message: h('div', { class: 'failed-uploads-notification' }, [
                    h('table', { class: 'failed-uploads-table' }, [
                        h('thead', {}, [
                            h('tr', {}, [
                                h('th', {}, '文件名'),
                                h('th', {}, '描述')
                            ])
                        ]),
                        h('tbody', {}, failedUploads.map(item =>
                            h('tr', {}, [
                                h('td', {}, item.fileName),
                                h('td', {}, item.description)
                            ])
                        ))
                    ])
                ]),
                type: 'error',
                duration: 0,
                showClose: true
            });
            // 保留失败项，方便改描述后重试
            const keep = pendingFiles.value.filter(item => item.status === 'error');
            const drop = pendingFiles.value.filter(item => item.status !== 'error');
            revokePendingPreviews(drop);
            pendingFiles.value = keep.map(item => ({ ...item, status: 'ready' as UploadItemStatus }));
        } else {
            ElNotification({
                title: '上传完成',
                message: `全部 ${queue.length} 个文件上传成功`,
                type: 'success',
                duration: 3000
            });
            isShowingModalUpload.value = false;
            resetUploadForm();
        }
        getFileList();
    } catch (err) {
        console.error('Batch upload failed:', err);
        ElMessage.error('获取上传凭证失败，请重试');
        queue.forEach(item => {
            if (item.status === 'ready' || item.status === 'uploading') {
                item.status = 'error';
            }
        });
    } finally {
        isUploading.value = false;
    }
}

const goDelete = (rows: TableData[]) => {
    const count = rows.length;
    ElMessageBox.confirm(`确定要删除选中的${count}张图片吗？`, '删除', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
    }).then(() => {
        const deletePromises = rows.map(row => 
            imageApi.delete({
                id: row.id,
                bucket: getSafeBucketName()
            })
        );
        
        Promise.all(deletePromises)
            .then(() => {
                ElNotification({
                    title: '删除成功',
                    type: 'success'
                });
                selectedRows.value = [];
                getFileList();
            })
            .catch(() => {
                ElNotification({
                    title: '删除失败',
                    type: 'error'
                });
            });
    });
};

const goBatchDelete = () => {
    if (selectedRows.value.length === 0) return;
    
    ElMessageBox.confirm(`确定要删除选中的${selectedRows.value.length}张图片吗？`, '批量删除', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
    }).then(() => {
        imageApi.batchDelete({
            ids: selectedRows.value.map(row => row.id),
            bucket: getSafeBucketName()
        }).then(() => {
            ElNotification({
                title: '批量删除成功',
                type: 'success'
            });
            selectedRows.value = [];
            getFileList();
        }).catch(() => {
            ElNotification({
                title: '批量删除失败',
                type: 'error'
            });
        });
    });
};

onMounted(() => {
    getFileList();
});

onUnmounted(() => {
    revokePendingPreviews();
});

const isShowingEditDialog = ref(false);
const editForm = ref({
    id: '',
    description: ''
});

const openEditDialog = (row: TableData) => {
    editForm.value = {
        id: row.id,
        description: row.description
    };
    isShowingEditDialog.value = true;
};

const submitEdit = () => {
    imageApi.update({
        id: editForm.value.id,
        description: editForm.value.description
    }).then(() => {
        ElNotification({
            title: '更新成功',
            type: 'success'
        });
        isShowingEditDialog.value = false;
        getFileList();
    }).catch(() => {
        ElNotification({
            title: '更新失败',
            type: 'error'
        });
    });
};


</script>

<style scoped lang="scss">
.thumbnail {
    width: 50px;
    height: 50px;
    overflow: hidden;
    margin: 0 auto;
    border-radius: 4px;
    border: 1px solid #dcdfe6;
    background: #f5f7fa;

    img {
        width: 100%;
        height: 100%;
        object-fit: cover;
    }
}

:deep(.failed-uploads-notification) {
    max-height: 300px;
    overflow-y: auto;
    
    .failed-uploads-table {
        width: 100%;
        border-collapse: collapse;
        margin-top: 10px;
        
        th, td {
            padding: 8px;
            text-align: left;
            border-bottom: 1px solid #ebeef5;
            font-size: 12px;
        }
        
        th {
            font-weight: bold;
            background-color: #f5f7fa;
        }
        
        tr:hover td {
            background-color: #f5f7fa;
        }
    }
}

.upload-dialog {
    display: flex;
    flex-direction: column;
    gap: 14px;
}

.upload-dropzone {
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 6px;
    min-height: 132px;
    padding: 20px 16px;
    border: 1.5px dashed #c0c4cc;
    border-radius: 10px;
    background: linear-gradient(180deg, #fafbfd 0%, #f3f5f8 100%);
    cursor: pointer;
    transition: border-color 0.2s, background 0.2s, box-shadow 0.2s;

    &:hover:not(.is-disabled),
    &.is-dragover {
        border-color: #409eff;
        background: linear-gradient(180deg, #f0f7ff 0%, #e8f3ff 100%);
        box-shadow: inset 0 0 0 1px rgba(64, 158, 255, 0.15);
    }

    &.is-disabled {
        cursor: not-allowed;
        opacity: 0.72;
    }

    &__input {
        display: none;
    }

    &__icon {
        color: #409eff;
        margin-bottom: 2px;
    }

    &__title {
        margin: 0;
        font-size: 14px;
        font-weight: 600;
        color: #303133;
    }

    &__hint {
        margin: 0;
        font-size: 12px;
        color: #909399;
    }
}

.upload-meta {
    display: flex;
    align-items: center;
    justify-content: space-between;
    font-size: 13px;
    color: #606266;
}

.upload-file-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
    gap: 12px;
    max-height: 360px;
    overflow-y: auto;
    padding: 2px;
}

.upload-file-card {
    display: flex;
    flex-direction: column;
    gap: 8px;

    &__thumb {
        position: relative;
        aspect-ratio: 1;
        border-radius: 8px;
        overflow: hidden;
        border: 1px solid #e4e7ed;
        background: #f5f7fa;

        img {
            width: 100%;
            height: 100%;
            object-fit: cover;
            display: block;
        }
    }

    &__remove {
        position: absolute;
        top: 6px;
        right: 6px;
        width: 22px;
        height: 22px;
        border: none;
        border-radius: 50%;
        background: rgba(0, 0, 0, 0.55);
        color: #fff;
        display: inline-flex;
        align-items: center;
        justify-content: center;
        cursor: pointer;
        padding: 0;

        &:hover {
            background: rgba(245, 108, 108, 0.95);
        }
    }

    &__mask {
        position: absolute;
        inset: 0;
        display: flex;
        align-items: center;
        justify-content: center;
        background: rgba(0, 0, 0, 0.45);
        color: #fff;
        font-size: 12px;
        font-weight: 600;

        &.is-success {
            background: rgba(103, 194, 58, 0.72);
        }

        &.is-error {
            background: rgba(245, 108, 108, 0.78);
        }
    }

    &.is-error .upload-file-card__thumb {
        border-color: #f56c6c;
    }

    &.is-success .upload-file-card__thumb {
        border-color: #67c23a;
    }
}

.edit-dialog-content {
    display: flex;
    flex-direction: column;
    gap: 20px;
    
    .edit-preview {
        width: 100%;
        height: 300px;
        border-radius: 4px;
        overflow: hidden;
        border: 1px solid #dcdfe6;
        background: #f5f7fa;
        display: flex;
        align-items: center;
        justify-content: center;
        
        img {
            max-width: 100%;
            max-height: 100%;
            object-fit: contain;
            padding: 8px;
        }
    }
    
    .edit-form-container {
        flex: 1;
        width: 100%;
        padding: 0 20px;
    }
}

.image-grid {
    padding: 10px;
    display: flex;
    flex-wrap: wrap;
    gap: 10px;

    .image-card {
        width: 160px;
        cursor: pointer;
        user-select: none;
        border: 1px solid #dcdfe6;
        border-radius: 4px;
        overflow: hidden;
        transition: all 0.2s;
        position: relative;

        .selection-order-badge {
            position: absolute;
            top: 8px;
            left: 8px;
            min-width: 22px;
            height: 22px;
            padding: 0 6px;
            border-radius: 11px;
            background: #409eff;
            color: #fff;
            font-size: 12px;
            font-weight: 600;
            line-height: 22px;
            text-align: center;
            z-index: 2;
            box-shadow: 0 1px 4px rgba(0, 0, 0, 0.2);
        }

        &:hover {
            border-color: #409eff;
        }

        .image-preview {
            width: 100%;
            height: 150px;
            overflow: hidden;
            border-bottom: 1px solid #dcdfe6;
            background-color: #f5f7fa;
            display: flex;
            align-items: center;
            justify-content: center;

            img {
                width: 100%;
                height: 100%;
                object-fit: contain;
                padding: 4px;
            }
        }

        .image-info {
            padding: 10px;

            .image-description {
                font-size: 13px;
                font-weight: 600;
                margin-bottom: 8px;
                white-space: nowrap;
                overflow: hidden;
                text-overflow: ellipsis;
            }

            .image-date {
                font-size: 12px;
                color: #909399;
            }

            .image-actions {
                margin-top: 10px;
                display: flex;
                align-items: center;
                justify-content: center;
                gap: 2px;

                .el-button {
                    margin-left: 0;
                    font-size: 12px;
                }
            }

            .copy-split {
                display: inline-flex;
                align-items: center;
                border-radius: 4px;

                .copy-split__main {
                    padding: 5px 4px 5px 6px !important;
                }

                .copy-split__caret {
                    padding: 5px 4px !important;
                    min-width: 20px;

                    .el-icon {
                        font-size: 11px;
                    }
                }
            }
        }

        &.selected {
            border-color: #409eff;
            background-color: #409eff;
            .image-info {
                .image-description {
                    color: white;
                }
                .image-date {
                    color: white;
                }
                .image-actions {
                    .el-button {
                        color: white !important;
                        .el-icon {
                            color: white !important;
                        }
                    }
                }
            }
        }
    }
}
</style>

<style lang="scss">
.copy-dropdown-compact {
    .el-dropdown-menu__item {
        font-size: 11px;
        line-height: 1.25;
        padding: 3px 10px;
        min-width: 0;
    }
}
</style>
