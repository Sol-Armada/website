<template>
    <v-app-bar v-if="$vuetify.display.mobile">
        <template v-slot:prepend>
            <v-app-bar-nav-icon @click.stop="drawer = !drawer"></v-app-bar-nav-icon>
        </template>

        <v-app-bar-title>Sol Armada</v-app-bar-title>
    </v-app-bar>

    <v-navigation-drawer :temporary="isMobile()" v-model="drawer">
        <v-list>
            <v-list-item>
                <v-sheet>
                    <v-skeleton-loader type="card" v-if="me == null"></v-skeleton-loader>
                    <v-row class="justify-center" v-else>
                        <v-col cols="auto">
                            <v-avatar size="100"
                                :image="'https://cdn.discordapp.com/avatars/' + me.id + '/' + me.avatar + '.png'"></v-avatar>
                        </v-col>
                        <v-col cols="auto">
                            <div class="text-center">
                                <div class="text-h5">{{ me.name }}</div>
                                <div>{{ me.rank.name }}</div>
                            </div>
                        </v-col>
                    </v-row>
                </v-sheet>
            </v-list-item>
            <v-divider></v-divider>
            <v-list-item link title="Home" to="/"></v-list-item>
            <v-list-item link title="Handbook" to="/handbook"></v-list-item>
            <div v-if="me && me.rank.id <= 3">
                <v-divider></v-divider>
                <v-list-item link title="Members" to="/members"></v-list-item>
                <v-list-item link title="Attendance Records" to="/attendance"></v-list-item>
            </div>
        </v-list>
        <template v-slot:append>
            <div class="pa-2">
                <v-btn block prepend-icon="mdi:mdi-logout" size="large" color="primary"
                    @click="appStore.logout()">Logout</v-btn>
            </div>
            <v-divider></v-divider>
            <div class="pa-2">
                <AppFooter />
            </div>
        </template>
    </v-navigation-drawer>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { storeToRefs } from 'pinia'
import { useAppStore } from '@/stores/app'

const appStore = useAppStore()
const { me } = storeToRefs(appStore)

const drawer = ref(!isMobile())

onMounted(() => {
    if (!appStore.me) {
        appStore.getMe()
    }
})

function isMobile() {
    return /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent)
}

</script>
