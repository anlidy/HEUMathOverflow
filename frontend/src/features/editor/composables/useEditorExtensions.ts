import { VueNodeViewRenderer } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import Placeholder from '@tiptap/extension-placeholder'
import Link from '@tiptap/extension-link'
import Image from '@tiptap/extension-image'
import CodeBlockLowlight from '@tiptap/extension-code-block-lowlight'
import { createLowlight } from 'lowlight'
// 只导入需要的语言，减小构建体积
import c from 'highlight.js/lib/languages/c'
import cpp from 'highlight.js/lib/languages/cpp'
import typescript from 'highlight.js/lib/languages/typescript'
import python from 'highlight.js/lib/languages/python'
import java from 'highlight.js/lib/languages/java'
import { InlineMath, BlockMath } from '@/features/editor/extensions'
import CodeBlockComponent from '@/features/editor/components/CodeBlockComponent.vue'

// Lowlight setup - 只注册需要的语言
const lowlight = createLowlight({ c, cpp, typescript, python, java })

/**
 * 获取共享的编辑器扩展配置
 * @param options 配置选项
 * @returns 扩展数组
 */
export function useEditorExtensions(options?: {
    placeholder?: string
    enablePlaceholder?: boolean
    enableCodeBlockShortcuts?: boolean
}) {
    const { placeholder, enablePlaceholder = false, enableCodeBlockShortcuts = false } = options || {}

    const extensions = [
        StarterKit.configure({
            codeBlock: false, // Disable default codeBlock to use lowlight
        }),
        ...(enablePlaceholder && placeholder
            ? [
                  Placeholder.configure({
                      placeholder,
                  }),
              ]
            : []),
        Link.configure({
            openOnClick: !enablePlaceholder, // 编辑模式下不允许点击链接，只读模式下允许
        }),
        Image,
        CodeBlockLowlight.extend({
            addNodeView() {
                return VueNodeViewRenderer(CodeBlockComponent)
            },
            ...(enableCodeBlockShortcuts
                ? {
                      addKeyboardShortcuts() {
                          return {
                              Enter: ({ editor }) => {
                                  // Ensure Enter creates a new line in code block
                                  if (editor.isActive('codeBlock')) {
                                      editor.commands.insertContent('\n')
                                      return true
                                  }
                                  return false
                              },
                          }
                      },
                  }
                : {}),
        }).configure({
            lowlight,
        }),
        InlineMath,
        BlockMath,
    ]

    return extensions
}

