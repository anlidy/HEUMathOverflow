import { useMessage, useDialog } from 'naive-ui'
import type { AppError } from '@/utils/errorHandler'
import { shouldShowDialog, shouldShowMessage } from '@/utils/errorHandler'

/**
 * 消息提示composable
 * 统一管理消息提示和错误弹窗
 */
export function useAppMessage() {
    const message = useMessage()
    const dialog = useDialog()

    /**
     * 显示成功消息
     */
    const showSuccess = (content: string, duration = 3000) => {
        message.success(content, { duration })
    }

    /**
     * 显示警告消息
     */
    const showWarning = (content: string, duration = 3000) => {
        message.warning(content, { duration })
    }

    /**
     * 显示信息消息
     */
    const showInfo = (content: string, duration = 3000) => {
        message.info(content, { duration })
    }

    /**
     * 显示错误消息
     */
    const showError = (content: string, duration = 3000) => {
        message.error(content, { duration })
    }

    /**
     * 处理AppError，根据错误严重程度自动选择提示方式
     */
    const handleError = (error: AppError) => {
        if (shouldShowDialog(error)) {
            // 重要错误使用弹窗
            dialog.error({
                title: '错误',
                content: error.message,
                positiveText: '确定',
                onPositiveClick: () => {},
            })
        } else if (shouldShowMessage(error)) {
            // 普通错误使用消息提示
            showError(error.message)
        }
        // SILENT错误不显示任何提示
    }

    /**
     * 显示错误弹窗（用于重要错误）
     */
    const showErrorDialog = (title: string, content: string) => {
        dialog.error({
            title,
            content,
            positiveText: '确定',
            onPositiveClick: () => {},
        })
    }

    /**
     * 显示确认对话框
     */
    const showConfirmDialog = (
        title: string,
        content: string,
        onConfirm: () => void | Promise<void>
    ) => {
        dialog.warning({
            title,
            content,
            positiveText: '确定',
            negativeText: '取消',
            onPositiveClick: async () => {
                await onConfirm()
            },
        })
    }

    return {
        showSuccess,
        showWarning,
        showInfo,
        showError,
        handleError,
        showErrorDialog,
        showConfirmDialog,
    }
}

