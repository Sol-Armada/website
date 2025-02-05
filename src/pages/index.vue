<!-- eslint-disable vue/multi-word-component-names -->
<template>
    <v-container fluid>
        <v-row class="justify-center">
            <v-col cols="3">
                <v-card :class="'bg-surface-' + ld + '-1'">
                    <v-row class="justify-center pa-5">
                        <v-col cols="12" class="text-center">
                            <v-avatar size="150"
                                :image="'https://cdn.discordapp.com/avatars/' + me.id + '/' + me.avatar + '.png'"></v-avatar>
                        </v-col>
                        <v-col cols="12">
                            <div class="text-center">
                                <div class="text-h5">{{ me.name }}</div>
                                <div class="text-subtitle">{{ me.rank.id >= 8 || me.rank.id == 0 ?
                                    me.affiliation : me.rank.name }}</div>
                            </div>
                        </v-col>
                        <v-divider></v-divider>
                        <v-col cols="12">
                            <div class="text-center">
                                <div class="text-h6">Events Attended</div>
                                {{ eventCount }}
                            </div>
                        </v-col>
                        <v-col cols="12" v-if="!me.isOfficer">
                            <div class="text-center">
                                <v-btn color="primary" @click="appStore.logout()">Logout</v-btn>
                            </div>
                        </v-col>
                    </v-row>
                </v-card>
            </v-col>
        </v-row>
    </v-container>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useTheme } from 'vuetify'
import { useAppStore } from '@/stores/app'
import { storeToRefs } from 'pinia'
import { useAttendanceStore } from '@/stores/attendance'

const appStore = useAppStore()
const attendanceStore = useAttendanceStore()

const { me } = storeToRefs(appStore)

const theme = useTheme()
const ld = ref('darken')
if (theme.current.value.dark) {
    ld.value = 'lighten'
}

const eventCount = ref(null)

onMounted(() => {
    attendanceStore.getAttendanceCount(me.value.id).then((res) => {
        eventCount.value = res
    })
})

</script>
<route lang="yaml">
meta:
    layout: default
</route>
