import { Node, mergeAttributes, nodeInputRule } from '@tiptap/core'
import { VueNodeViewRenderer } from '@tiptap/vue-3'
import MathComponent from './MathComponent.vue'

export const MathExtension = Node.create({
  name: 'mathComponent',

  group: 'block',

  atom: true,

  addAttributes() {
    return {
      latex: {
        default: 'E = mc^2',
      },
    }
  },

  parseHTML() {
    return [
      {
        tag: 'math-component',
      },
    ]
  },

  renderHTML({ HTMLAttributes }) {
    return ['math-component', mergeAttributes(HTMLAttributes)]
  },

  addNodeView() {
    return VueNodeViewRenderer(MathComponent as any)
  },

  addInputRules() {
    return [
      nodeInputRule({
        find: /^\$\$\s$/,
        type: this.type,
      }),
    ]
  },
})

