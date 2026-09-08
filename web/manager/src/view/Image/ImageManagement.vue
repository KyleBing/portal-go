<template>
    <Container>
        <Toolbar>
            <template #left>
                <ElForm inline class="tool-bar" @submit.prevent="getFileList">
                    <ElFormItem label="">
                        <ElRadioGroup v-model="currentBucket">
                            <ElRadioButton :value="item"
                                           :label="item"
                                           v-for="item in BucketQiniuMap.keys()" :key="item"/>
                        </ElRadioGroup>
                    </ElFormItem>
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
                                class="clipboard"
                                :data-clipboard="getBatchClipboardText(copyOption)">
                                {{ copyOption.label }}
                            </ElDropdownItem>
                        </ElDropdownMenu>
                    </template>
                </ElDropdown>
                <ElCheckbox v-model="copyAsMarkdown">复制为 Markdown 图片</ElCheckbox>
            </template>
        </Toolbar>

        <Content padding="0">
            <div class="image-grid" v-loading="isLoading">
                <div v-for="item in tableData" 
                     :key="item.id" 
                     class="image-card"
                     :class="{ 'selected': selectedRows.includes(item) }"
                     @click="toggleSelection(item)">
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
                            <ElDropdown trigger="hover">
                                <ElButton size="small" type="info" icon="CopyDocument" text plain>
                                    <ElIcon class="el-icon--right"><ArrowDown /></ElIcon>
                                </ElButton>
                                <template #dropdown>
                                    <ElDropdownMenu>
                                        <ElDropdownItem v-for="(resolution, key) in ImageResolutions" 
                                                       :key="key"
                                                       class="clipboard"
                                                       :data-clipboard="`${currentBucketInfo?.baseUrl}/${item.id}-${resolution.name}`">
                                            {{ resolution.name }} ({{ resolution.width }}x{{ resolution.height }})
                                        </ElDropdownItem>
                                        <ElDropdownItem class="clipboard"
                                                       :data-clipboard="`${currentBucketInfo?.baseUrl}/${item.id}`">
                                            原图
                                        </ElDropdownItem>
                                        <ElDropdownItem class="clipboard"
                                                       :data-clipboard="item.id">
                                            图片标识（无前缀）
                                        </ElDropdownItem>
                                    </ElDropdownMenu>
                                </template>
                            </ElDropdown>
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
                    <ElForm size="small" label-width="80px">
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
            width="600px"
            center>
            <ElForm size="small" label-width="80px">
                <ElFormItem label="选择文件">
                    <input type="file" ref="inputUpload" @change="inputFileChange" multiple accept="image/*"/>
                </ElFormItem>
                <ElFormItem label="文件描述">
                    <div v-for="file in formImage.files" :key="file.name" class="file-description-item">
                        <div class="file-preview">
                            <img :src="previewUrls[file.name]" alt="preview" class="preview-image">
                        </div>
                        <div class="file-info">
                            <div class="file-name">{{ file.name }}</div>
                            <ElInput type="text" v-model="formImage.fileDescriptions[file.name]" placeholder="请输入图片描述"></ElInput>
                        </div>
                    </div>
                </ElFormItem>
                <ElFormItem label="BUCKET">
                    <ElRadioGroup v-model="formImage.bucket">
                        <ElRadio :label="item" v-for="item in BucketQiniuMap.keys()" :key="item"></ElRadio>
                    </ElRadioGroup>
                </ElFormItem>
            </ElForm>
            <template #footer center class="dialog-footer">
                <ElButton @click="isShowingModalUpload = false">取 消</ElButton>
                <ElButton icon="Upload" type="primary" @click="fileUpload">确 定</ElButton>
            </template>
        </ElDialog>

    </Container>
</template>

<script setup lang="ts">
import {ref, computed, onMounted, watch, onUnmounted, h} from 'vue';
import { useProjectStore } from "@/pinia";
import imageApi, { ImageQiniuRequest } from "@/api/imageQiniuApi.ts";
import * as qiniu from 'qiniu-js';
import ClipboardJS from 'clipboard';
import Container from "@/layout/Container.vue";
import Toolbar from "@/layout/Toolbar.vue";
import FooterPagination from "@/layout/FooterPagination.vue";
import {BucketQiniuMap} from "@/configProject.ts";
import Content from "@/layout/Content.vue";
import TableListDate from "@/components/TableListDate.vue";
import imageQiniuApi from "@/api/imageQiniuApi.ts";
import { ImageResolutions, } from '@/config/imageResolutions';
import { ArrowDown } from '@element-plus/icons-vue'
import {ElMessage, ElMessageBox, ElNotification} from "element-plus";
import {dateFormatter} from "@/utility.ts";


interface ImageForm {
    files: File[];
    fileDescriptions: { [key: string]: string };
    bucket: string;
}

interface TableData {
    id: string;
    description: string;
    type: string;
    date_create: string;
}

const projectStore = useProjectStore();
const isLoading = ref(false);
const tableData = ref<TableData[]>([]);
const isShowingModalUpload = ref(false);
const searchKeyword = ref('');
const inputUpload = ref<HTMLInputElement | null>(null);


const currentBucket = ref('diary-container')
watch(currentBucket, () => {
    getFileList()
})
const currentBucketInfo = computed(() => {
    return BucketQiniuMap.get(currentBucket.value)
})

// Fix TypeScript errors for image URLs and bucket info
function getSafeImageUrl(id: string, resolution: string) {
    return currentBucketInfo.value ? `${currentBucketInfo.value.baseUrl}/${id}-${resolution}` : '';
}

function getSafeBucketName() {
    return currentBucketInfo.value?.bucketName || 'diary-container';
}

const formImage = ref<ImageForm>({
    files: [],
    fileDescriptions: {},
    bucket: 'diary-container'
});

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

const previewUrls = ref<{ [key: string]: string }>({});

const selectedRows = ref<TableData[]>([]);
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

// Track failed uploads
const failedUploads = ref<{fileName: string; description: string}[]>([]);

const handleSelectionChange = (selection: TableData[]) => {
    selectedRows.value = selection;
};

const toggleSelection = (item: TableData) => {
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
        })
        .catch(() => {
            isLoading.value = false;
        });
};

const addNewImage = () => {
    isShowingModalUpload.value = true;
};


function inputFileChange(event: Event) {
    const input = event.target as HTMLInputElement;
    if (input.files && input.files.length > 0) {
        const files = Array.from(input.files);
        const validFiles = files.filter(file => {
            if (!/image\/.*/.test(file.type)) {
                ElMessage.warning(`文件 ${file.name} 不是图片文件`);
                return false;
            }
            if (file.size > 1024 * 1024 * 10) {
                ElMessage.warning(`文件 ${file.name} 大小超过10MB`);
                return false;
            }
            return true;
        });
        
        // Clean up old preview URLs
        Object.values(previewUrls.value).forEach(url => URL.revokeObjectURL(url));
        previewUrls.value = {};
        
        // Create new preview URLs
        validFiles.forEach(file => {
            previewUrls.value[file.name] = URL.createObjectURL(file);
        });
        
        formImage.value.files = validFiles;
        // Initialize descriptions for new files
        validFiles.forEach(file => {
            if (!formImage.value.fileDescriptions[file.name]) {
                // Use file name without extension as default description
                const fileName = file.name.split('.').slice(0, -1).join('.');
                formImage.value.fileDescriptions[file.name] = fileName || file.name;
            }
        });
    }
}

// 上传文件
function fileUpload() {
    if (formImage.value.files.length === 0) {
        ElMessage.warning('请先选择文件');
        return;
    }

    // Reset failed uploads
    failedUploads.value = [];
    
    const uploadNextFile = (index: number) => {
        if (index >= formImage.value.files.length) {
            // All files uploaded
            isShowingModalUpload.value = false;
            
            // Show failed uploads notification if any
            if (failedUploads.value.length > 0) {
                console.log('Failed uploads:', failedUploads.value);
                ElNotification({
                    title: `${failedUploads.value.length} 个文件上传失败`,
                    message: h('div', { class: 'failed-uploads-notification' }, [
                        h('table', { class: 'failed-uploads-table' }, [
                            h('thead', {}, [
                                h('tr', {}, [
                                    h('th', {}, '文件名'),
                                    h('th', {}, '描述')
                                ])
                            ]),
                            h('tbody', {}, failedUploads.value.map(item => 
                                h('tr', {}, [
                                    h('td', {}, item.fileName),
                                    h('td', {}, item.description)
                                ])
                            ))
                        ])
                    ]),
                    type: 'error',
                    duration: 0, // Don't auto dismiss
                    showClose: true
                });
            } else {
                console.log('No failed uploads');
                ElNotification({
                    title: '上传完成',
                    message: `所有 ${formImage.value.files.length} 个文件上传成功`,
                    type: 'success',
                    duration: 3000
                });
            }
            
            formImage.value = {
                files: [],
                fileDescriptions: {},
                bucket: getSafeBucketName()
            };
            getFileList();
            return;
        }

        const currentFile = formImage.value.files[index];
        
        imageQiniuApi
            .getUploadToken({
                bucket: formImage.value.bucket,
                hahaha: true
            })
            .then(res => {
                const observer = {
                    next: (res: any) => {
                        // Upload progress
                    },
                    error: (err: any) => {
                        console.error('Upload error for file:', currentFile.name, err);
                        // Record failed upload
                        failedUploads.value.push({
                            fileName: currentFile.name,
                            description: formImage.value.fileDescriptions[currentFile.name] || ''
                        });
                        console.log('Added to failed uploads:', failedUploads.value);
                        ElMessage.error(`文件 ${currentFile.name} 上传失败`);
                        uploadNextFile(index + 1);
                    },
                    complete: (res: any) => {
                        fileNewSubmit({
                            id: res.key,
                            type: 'image',
                            bucket: formImage.value.bucket,
                            description: formImage.value.fileDescriptions[currentFile.name],
                        });
                        uploadNextFile(index + 1);
                    }
                };

                const observable = qiniu.upload(currentFile, null, res.data, {}, {});
                observable.subscribe(observer);
            })
            .catch(err => {
                console.error('Token error for file:', currentFile.name, err);
                // Record failed upload due to token error
                failedUploads.value.push({
                    fileName: currentFile.name,
                    description: formImage.value.fileDescriptions[currentFile.name] || ''
                });
                console.log('Added to failed uploads (token error):', failedUploads.value);
                ElMessage.error(`获取上传凭证失败`);
                uploadNextFile(index + 1);
            });
    };

    uploadNextFile(0);
}
// 新增
function fileNewSubmit(content: ImageQiniuRequest) {
    imageQiniuApi
        .add(content)
        .then((res: any) => {
            ElNotification({
                title: res.message,
                position: 'top-right',
                type: 'success',
                onClose() {
                }
            })
            isShowingModalUpload.value = false
            getFileList()
        })
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

let clipboard: ClipboardJS | null = null;
onMounted(() => {
    getFileList();
    // 绑定剪贴板操作方法
    clipboard = new ClipboardJS('.clipboard', {
        text: trigger => {
            return trigger.getAttribute('data-clipboard') || '';
        },
    });
    clipboard.on('success', () => {
        ElMessage.success('复制成功');
    });
    clipboard.on('error', () => {
        ElMessage.error('复制失败，请手动复制');
    });
});

onUnmounted(() => {
    if (clipboard) {
        clipboard.destroy();
    }
    // Clean up preview URLs
    Object.values(previewUrls.value).forEach(url => URL.revokeObjectURL(url));
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

.file-description-item {
    margin-bottom: 15px;
    display: flex;
    gap: 15px;
    align-items: flex-start;
    
    .file-preview {
        width: 120px;
        min-width: 120px;
        border-radius: 4px;
        overflow: hidden;
        border: 1px solid #dcdfe6;
        background: #f5f7fa;
        display: flex;
        align-items: center;
        justify-content: center;
        
        .preview-image {
            max-width: 100%;
            max-height: 120px;
            object-fit: contain;
        }
    }
    
    .file-info {
        flex: 1;
        min-width: 0;
        
        .file-name {
            font-size: 12px;
            color: #666;
            margin-bottom: 8px;
            word-break: break-all;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }
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
                text-align: center;
                .el-button {
                    margin-left: 5px;
                    font-size: 12px;
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
