<script setup lang="ts">
import type { BrandIconName } from '../../../../design/brand/envarium-v1/icons'

withDefaults(defineProps<{ name: BrandIconName; size?: number }>(), { size: 18 })
</script>

<template>
  <svg
    :width="size"
    :height="size"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="1.8"
    stroke-linecap="round"
    stroke-linejoin="round"
    aria-hidden="true"
    focusable="false"
  >
    <use :href="`/brand/interface.svg#${name}`" />
  </svg>
</template>
