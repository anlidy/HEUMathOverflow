import { Node, mergeAttributes, nodeInputRule } from '@tiptap/core'
import { VueNodeViewRenderer } from '@tiptap/vue-3'
import BlockMathComponent from './BlockMathComponent.vue'

/**
 * 块级数学公式扩展
 * 使用 $$空格 或 /math 触发
 */
export const BlockMath = Node.create({
    name: 'blockMath',

    group: 'block',

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
                tag: 'div[data-type="block-math"]',
                getAttrs: (node) => {
                    if (typeof node === 'string') return {}
                    return {
                        latex: node.getAttribute('data-latex') || '',
                    }
                },
            },
            // 兼容旧的 math-component 标签
            {
                tag: 'math-component',
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
            'div',
            mergeAttributes(HTMLAttributes, {
                'data-type': 'block-math',
                'data-latex': node.attrs.latex,
            }),
        ]
    },

    addNodeView() {
        return VueNodeViewRenderer(BlockMathComponent)
    },

    addInputRules() {
        return [
            // $$ + 空格触发
            nodeInputRule({
                find: /^\$\$\s$/,
                type: this.type,
            }),
            // /math 触发
            nodeInputRule({
                find: /^\/math\s$/,
                type: this.type,
            }),
        ]
    },
})
