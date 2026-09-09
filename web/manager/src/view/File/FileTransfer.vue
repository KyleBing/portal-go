<template>
    <Container>
        <Toolbar :is-show-search-bar="false"/>
        <Content>
            <div class="xfer">
                <ElAlert
                    type="info"
                    :closable="false"
                    title="端对端直连传文件"
                    description="文件只经 WebRTC 点对点传输，不经过本站服务器流量。跨公网可能因 NAT 无法直连，请尽量同一 Wi‑Fi。"
                    class="mb"
                />

                <ElCard shadow="never" class="card">
                    <template #header>
                        <div class="card-head">
                            <span>房间</span>
                            <ElTag :type="statusTag">{{ statusText }}</ElTag>
                        </div>
                    </template>
                    <div class="row">
                        <ElButton type="primary" @click="createRoom" :disabled="!authed">创建房间</ElButton>
                        <ElInput v-model="joinCode" placeholder="输入对方房间码" style="max-width:180px" maxlength="8"/>
                        <ElButton @click="joinRoom" :disabled="!authed || !joinCode">加入</ElButton>
                        <ElButton text type="danger" @click="reset">断开</ElButton>
                    </div>
                    <p v-if="room" class="room-line">
                        房间码 <code class="code">{{ room }}</code>
                        <ElButton link type="primary" @click="copyRoom">复制</ElButton>
                        <span class="muted">peer {{ peerId }}</span>
                    </p>
                </ElCard>

                <ElCard shadow="never" class="card">
                    <template #header>发送文件</template>
                    <input type="file" ref="fileInput" @change="onPick"/>
                    <div class="row mt">
                        <ElButton type="success" :disabled="status !== 'ready' || !picked" @click="send">发送给对方</ElButton>
                        <span v-if="picked" class="muted">{{ picked.name }}（{{ pretty(picked.size) }}）</span>
                    </div>
                    <ElProgress v-if="progress" :percentage="progress.percent" :stroke-width="10" class="mt"/>
                </ElCard>

                <ElCard v-if="received.length" shadow="never" class="card">
                    <template #header>已接收</template>
                    <div v-for="(f, i) in received" :key="i" class="recv">
                        <span>{{ f.name }}（{{ pretty(f.blob.size) }}）</span>
                        <ElButton size="small" type="primary" @click="saveReceived(f)">保存</ElButton>
                    </div>
                </ElCard>
            </div>
        </Content>
    </Container>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import Container from '@/layout/Container.vue'
import Toolbar from '@/layout/Toolbar.vue'
import Content from '@/layout/Content.vue'
import { getAuthorization } from '@/utility'
import { PeerTransfer, type TransferProgress, type TransferStatus } from '@/utils/webrtcTransfer'

const auth = getAuthorization()
const authed = computed(() => !!(auth?.token && auth?.uid))

const transfer = new PeerTransfer()
const status = ref<TransferStatus>('idle')
const statusDetail = ref('')
const room = ref('')
const peerId = ref('')
const joinCode = ref('')
const picked = ref<File | null>(null)
const progress = ref<TransferProgress | null>(null)
const received = ref<{ name: string; mime: string; blob: Blob }[]>([])
const fileInput = ref<HTMLInputElement | null>(null)

const statusText = computed(() => statusDetail.value || status.value)
const statusTag = computed(() => {
    switch (status.value) {
        case 'ready': return 'success'
        case 'error': return 'danger'
        case 'sending':
        case 'receiving': return 'warning'
        default: return 'info'
    }
})

function pretty(n: number) {
    if (n < 1024) return `${n} B`
    if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`
    return `${(n / 1024 / 1024).toFixed(2)} MB`
}

function wire() {
    transfer.onStatus = (s, d) => {
        status.value = s
        statusDetail.value = d || ''
        if (s === 'error' && d) ElMessage.warning(d)
    }
    transfer.onProgress = (p) => { progress.value = p }
    transfer.onRoom = (r, p) => { room.value = r; peerId.value = p }
    transfer.onFile = (f) => {
        received.value.unshift(f)
        ElMessage.success(`收到文件 ${f.name}`)
    }
}

function createRoom() {
    if (!authed.value) return
    transfer.connect(auth!.token, auth!.uid)
    setTimeout(() => transfer.createRoom(), 200)
}

function joinRoom() {
    if (!authed.value) return
    transfer.connect(auth!.token, auth!.uid)
    setTimeout(() => transfer.joinRoom(joinCode.value), 200)
}

function onPick(e: Event) {
    const t = e.target as HTMLInputElement
    picked.value = t.files?.[0] || null
}

function send() {
    if (picked.value) transfer.sendFile(picked.value)
}

function saveReceived(f: { name: string; blob: Blob }) {
    const a = document.createElement('a')
    const href = URL.createObjectURL(f.blob)
    a.href = href
    a.download = f.name
    a.click()
    URL.revokeObjectURL(href)
}

async function copyRoom() {
    try {
        await navigator.clipboard.writeText(room.value)
        ElMessage.success('房间码已复制')
    } catch {
        ElMessage.error('复制失败')
    }
}

function reset() {
    transfer.disconnect()
    status.value = 'idle'
    statusDetail.value = ''
    room.value = ''
    peerId.value = ''
    progress.value = null
}

onMounted(() => {
    wire()
    if (authed.value) transfer.connect(auth!.token, auth!.uid)
})

onBeforeUnmount(() => transfer.disconnect())
</script>

<style scoped lang="scss">
.xfer { max-width: 720px; }
.mb { margin-bottom: 14px; }
.mt { margin-top: 12px; }
.card { margin-bottom: 14px; }
.card-head { display: flex; justify-content: space-between; align-items: center; }
.row { display: flex; flex-wrap: wrap; gap: 10px; align-items: center; }
.room-line { margin: 12px 0 0; font-size: 14px; }
.code {
    font-size: 20px;
    font-weight: 700;
    letter-spacing: 0.12em;
    margin: 0 8px;
}
.muted { color: #86909c; font-size: 12px; }
.recv {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 8px 0;
    border-bottom: 1px solid #f0f2f5;
}
</style>
