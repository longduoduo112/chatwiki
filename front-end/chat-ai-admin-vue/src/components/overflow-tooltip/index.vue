<script>
import {
  Comment,
  cloneVNode,
  computed,
  defineComponent,
  h,
  nextTick,
  onBeforeUnmount,
  onMounted,
  ref,
  watch
} from 'vue'
import { Tooltip as ATooltip } from 'ant-design-vue'

export default defineComponent({
  name: 'OverflowTooltip',
  inheritAttrs: false,
  props: {
    title: {
      type: [String, Number],
      default: ''
    },
    tooltipWidth: {
      type: [String, Number],
      default: undefined
    }
  },
  setup(props, { attrs, slots }) {
    const triggerRef = ref(null)
    const isOverflowing = ref(false)
    let resizeObserver = null

    const tooltipAttrs = computed(() => {
      const rest = { ...attrs }
      delete rest.class
      delete rest.style
      delete rest.overlayStyle
      delete rest.overlayInnerStyle
      return rest
    })

    const normalizedTooltipWidth = computed(() => {
      if (props.tooltipWidth === undefined || props.tooltipWidth === null || props.tooltipWidth === '') {
        return undefined
      }

      return typeof props.tooltipWidth === 'number' ? `${props.tooltipWidth}px` : props.tooltipWidth
    })

    const overlayStyle = computed(() => ({
      ...(attrs.overlayStyle || {}),
      ...(normalizedTooltipWidth.value
        ? {
            width: normalizedTooltipWidth.value,
            maxWidth: normalizedTooltipWidth.value
          }
        : {})
    }))

    const overlayInnerStyle = computed(() => ({
      ...(attrs.overlayInnerStyle || {}),
      ...(normalizedTooltipWidth.value
        ? {
            width: '100%',
            maxWidth: '100%'
          }
        : {})
    }))

    const getTriggerElement = () => triggerRef.value?.$el || triggerRef.value

    const updateOverflow = () => {
      nextTick(() => {
        const el = getTriggerElement()
        if (!el) return

        isOverflowing.value = el.scrollHeight > el.clientHeight + 1 || el.scrollWidth > el.clientWidth + 1
      })
    }

    const renderTrigger = () => {
      const children = (slots.default?.() || []).filter((child) => child.type !== Comment)
      if (children.length === 1) {
        return cloneVNode(
          children[0],
          {
            ref: triggerRef,
            class: attrs.class,
            style: attrs.style
          },
          true
        )
      }

      return h(
        'div',
        {
          ref: triggerRef,
          class: attrs.class,
          style: attrs.style
        },
        children
      )
    }

    onMounted(() => {
      nextTick(() => {
        const el = getTriggerElement()
        if (!el) return

        updateOverflow()
        if (typeof ResizeObserver !== 'undefined') {
          resizeObserver = new ResizeObserver(updateOverflow)
          resizeObserver.observe(el)
        }
      })
    })

    watch(() => props.title, updateOverflow)

    onBeforeUnmount(() => {
      resizeObserver?.disconnect()
    })

    return () =>
      h(
        ATooltip,
        {
          ...tooltipAttrs.value,
          title: isOverflowing.value ? props.title : undefined,
          overlayStyle: overlayStyle.value,
          overlayInnerStyle: overlayInnerStyle.value
        },
        {
          default: renderTrigger
        }
      )
  }
})
</script>
