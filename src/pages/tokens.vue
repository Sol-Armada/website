<!-- eslint-disable vue/multi-word-component-names -->
<template>
  <v-container fluid>
    <v-row class="justify-center">

      {{ tokens }}

    </v-row>
  </v-container>
</template>

<script setup>
import { onMounted } from 'vue'
import { useTokensStore } from '@/stores/tokens'
import { storeToRefs } from 'pinia'

const tokensStore = useTokensStore()

const { tokens } = storeToRefs(tokensStore)

onMounted(async () => {
  await tokensStore.getTokenRecords()
})

</script>
<route lang="yaml">
meta:
  layout: plain
</route>
