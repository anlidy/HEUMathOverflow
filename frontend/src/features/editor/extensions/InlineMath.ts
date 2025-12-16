import { Node, mergeAttributes, InputRule } from '@tiptap/core'
import { VueNodeViewRenderer } from '@tiptap/vue-3'
import InlineMathComponent from './InlineMathComponent.vue'

/**
 * 行内数学公式扩展
 * 使用 $...$ 语法输入行内公式
 * 例如：输入 $E=mc^2$ 会转换为行内公式
 */
export const InlineMath = Node.create({
    name: 'inlineMath',

    group: 'inline',

    inline: true,

    atom: true,

    addAttributes() {
        return {
            latex: {
                default: '',
            },
        }
    },

    parseHTML() {
        return [
            {
                tag: 'span[data-type="inline-math"]',
                getAttrs: (node) => {
                    if (typeof node === 'string') return {}
                    return {
                        latex: node.getAttribute('data-latex') || '',
                    }
                },
            },
        ]
    },

    renderHTML({ HTMLAttributes, node }) {
        return [
            'span',
            mergeAttributes(HTMLAttributes, {
                'data-type': 'inline-math',
                'data-latex': node.attrs.latex,
            }),
        ]
    },

    addNodeView() {
        return VueNodeViewRenderer(InlineMathComponent)
    },

    addInputRules() {
        const type = this.type

        return [
            // 匹配 $...$ 格式 - 使用更简单的方式
            new InputRule({
                find: /(?:^|[^$])\$([^$]+)\$$/,
                handler: ({ state, range, match }) => {
                    const latex = match[1]?.trim()
                    if (!latex) return null

                    // 调整 range，因为正则可能匹配了前面的非 $ 字符
                    const start = match[0].startsWith('$') ? range.from : range.from + 1
                    const { tr } = state

                    tr.replaceWith(start, range.to, type.create({ latex }))
                    return
                },
            }),
        ]
    },

    addKeyboardShortcuts() {
        return {
            // Ctrl/Cmd + M 插入行内公式
            'Mod-m': () => {
                return this.editor
                    .chain()
                    .focus()
                    .insertContent({
                        type: this.name,
                        attrs: { latex: '' },
                    })
                    .run()
            },
        }
    },
})
