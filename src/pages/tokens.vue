<!-- eslint-disable vue/multi-word-component-names -->
<template>
  <v-container fluid>
    <v-row class="justify-center">

      <v-data-table :items="tokens" :headers="headers" v-model:sort-by.sync="sortBy"></v-data-table>

    </v-row>
  </v-container>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useTokensStore } from '@/stores/tokens'
import { storeToRefs } from 'pinia'

const tokensStore = useTokensStore()

const { tokens } = storeToRefs(tokensStore)

const sortBy = ref([{ key: 'createdAt', order: 'desc' }]);

const headers = [
  { title: 'Member', value: 'member' },
  { title: 'Giver', key: 'giver', value: item => item.giver || 'Bot' },
  { title: 'Amount', value: 'amount' },
  { title: 'Reason', value: 'reason' },
  { title: 'Attendance ID', value: 'attendanceId' },
  { title: 'Comment', value: 'comment' },
  {
    title: 'Created Date', key: 'createdAt', value: item => {
      const date = new Date(item.createdAt);
      return isNaN(date.getTime()) ? new Date(item.created_at).toLocaleString() : date.toLocaleString();
    },
  },
]

onMounted(async () => {
  await tokensStore.getTokenRecords(0)
})

</script>
<route lang="yaml">
meta:
  layout: default
  requiresOfficer: true
</route>
