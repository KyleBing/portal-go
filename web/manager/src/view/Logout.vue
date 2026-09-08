<template>

</template>

<script setup lang="ts">

import {useRouter} from "vue-router";
import {deleteAuthorization} from "@/utility.ts";
const router = useRouter()
router.push('/login')

deleteAuthorization()

</script>

<style scoped>

</style>
