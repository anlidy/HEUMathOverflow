<script setup lang="ts">
import { NMessageProvider, NDialogProvider } from 'naive-ui'
</script>

<template>
    <n-message-provider>
        <n-dialog-provider>
            <router-view />
        </n-dialog-provider>
    </n-message-provider>
</template>

<style scoped></style>
