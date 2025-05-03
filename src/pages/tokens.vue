<!-- eslint-disable vue/multi-word-component-names -->
<template>
  <v-container fluid>
    <v-row class="justify-center">
      <v-col cols="6">
        <v-card class="bg-card-on-surface">
          <v-card-title>Tokens</v-card-title>
          <v-divider></v-divider>

          <!-- <v-data-table :items="tokens" :headers="headers" v-model:sort-by.sync="sortBy"></v-data-table> -->

          <v-card class="bg-card-on-surface">
            <template v-slot:text>
              <v-text-field v-model="tokensListSearch" label="Search" prepend-inner-icon="mdi:mdi-magnify"
                variant="outlined" hide-details single-line></v-text-field>
            </template>
            <v-container fluid :style="{ height: '100%' }">
              <v-data-table id="tokens" class="bg-card-on-surface" :items="tokens" :disable-items-per-page=true
                :headers="headers" density="compact" :search="tokensListSearch" :loading="loading"
                :itemsPerPageOptions="[12]" :loading-text="loadingText" color="white" v-model:sort-by.sync="sortBy"
                v-touch="{
                  left: () => swipe('Left'),
                  right: () => swipe('Right')
                }">
                <template v-slot:item="{ item }">
                  <tr :id="item.id" @click="tokenDetail = item">
                    <td width="33.33%">
                      <v-col cols="12">
                        <v-card>
                          <v-card-text>{{ item.member_id }}</v-card-text>
                        </v-card>
                      </v-col>
                    </td>
                    <td width="25%">
                      <v-col cols="12">
                        <v-card>
                          <v-card-text>{{ item.amount }}</v-card-text>
                        </v-card>
                      </v-col>
                    </td>
                    <td>
                      <v-col cols="12">
                        <v-card>
                          <v-card-text>{{ item.reason }}</v-card-text>
                        </v-card>
                      </v-col>
                    </td>
                    <td>
                      <v-col cols="12">
                        <v-card>
                          <v-card-text>{{ item.attendance_id }}</v-card-text>
                        </v-card>
                      </v-col>
                    </td>
                  </tr>
                </template>
              </v-data-table>
            </v-container>
          </v-card>

        </v-card>
      </v-col>
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

const tokensListSearch = ref('')

onMounted(async () => {
  await tokensStore.getTokenRecords(0)
})

</script>
<route lang="yaml">
meta:
  layout: default
  requiresOfficer: true
</route>
