<!-- eslint-disable vue/multi-word-component-names -->
<template>
    <v-container class="fill-height">
        <v-responsive class="align-centerfill-height mx-auto">

            <v-row class="justify-center">
                <v-col md="6">

                    <v-card class="pa-6 rounded-lg bg-surface-lighten-1">
                        <v-img class="mb-4" height="300" min-width="300" src="@/assets/logo-blue.png"
                            v-if="theme.name.value == 'light'" />
                        <v-img class="mb-4" height="300" min-width="300" src="@/assets/logo-white.png" v-else />

                        <div class="text-center">

                            <div class="py-2" />

                            <v-form>
                                <v-row class="justify-center">
                                    <v-col md="12">
                                        <v-text-field name="rsi_handle" label="RSI Handle" v-model="rsiHandle"
                                            @input="debounce()"
                                            :hint="'https://robersspaceindustries.com/citizens/' + rsiHandle"
                                            persistent-hint clearable :loading="checkingRSIHandle"
                                            :validation-value="rsiHandleValid"
                                            :rules="[(v) => !!v || 'RSI Handle is required and needs to be valid']"></v-text-field>
                                    </v-col>
                                </v-row>

                                <v-row class="justify-center">
                                    <v-col md="12">
                                        <v-select label="How did you find us?" required v-model="foundBy"
                                            :items="[{ title: 'RSI Website/Spectrum', value: 'rsi_website' }, { title: 'Recruited by a member', value: 'recruited' }, { title: 'Other', value: 'other' }]"></v-select>
                                    </v-col>
                                </v-row>

                                <v-row v-if="foundBy && foundBy == 'recruited' && !me.onboarded" class="justify-center">
                                    <v-col cols="12">
                                        <v-combobox label="Who recruited you?" :items="members" v-model="recruitedBy"
                                            :required="foundBy && foundBy.value == 'recruited'"
                                            :error="recruitedBy && recruitedBy.title == rsiHandle || !recruitedBy || !members.includes(recruitedBy)"
                                            :error-messages="recruitedBy && recruitedBy.title == rsiHandle ? 'You can\'t recruit yourself' : null"></v-combobox>
                                    </v-col>
                                </v-row>
                                <v-row v-else-if="foundBy && foundBy == 'recruited' && me.recruitedBy && me.onboarded"
                                    class="justify-center">
                                    <v-col cols="12">
                                        <v-combobox v-model="me.recruitedBy.name" disabled></v-combobox>
                                    </v-col>
                                </v-row>

                                <v-row v-if="foundBy && foundBy == 'other'" class=" justify-center">
                                    <v-col cols="12">
                                        <v-textarea label="What is your story?" v-model="me.other"></v-textarea>
                                    </v-col>
                                </v-row>

                                <v-row class="justify-center" v-if="!isMobile()">
                                    <v-col cols="12" md="6">
                                        <h5>Years playing Star Citizen (Round up)</h5>
                                        <v-number-input :reverse="false" controlVariant="split" label=""
                                            :hideInput="false" :inset="false" v-model="me.playTime" min=0
                                            max=10></v-number-input>
                                    </v-col>
                                    <v-col cols="12" md="6">
                                        <h5>How old are you?</h5>
                                        <v-number-input :reverse="false" controlVariant="split" label=""
                                            :hideInput="false" :inset="false" v-model="me.age"></v-number-input>
                                    </v-col>
                                </v-row>

                                <v-row class="justify-center text-left" v-else>
                                    <v-col cols="12" md="6">
                                        <h5>Years playing Star Citizen (Round up)</h5>
                                        <v-slider v-model="me.playTime" step=1 :min=0 :max=10>
                                            <template v-slot:append>
                                                <v-text-field v-model="me.playTime" density="compact"
                                                    style="width: 70px" type="number" hide-details
                                                    single-line></v-text-field>
                                            </template>
                                        </v-slider>
                                    </v-col>
                                    <v-col cols="12" md="6">
                                        <h5>How old are you?</h5>
                                        <v-slider v-model="me.age" step=1 :max=99>
                                            <template v-slot:append>
                                                <v-text-field v-model="me.age" density="compact" style="width: 70px"
                                                    type="number" hide-details single-line></v-text-field>
                                            </template>
                                        </v-slider>
                                    </v-col>
                                </v-row>

                                <v-row class="justify-center">
                                    <v-col cols="12">
                                        <v-select chips multiple clearable v-model="chosenGameplay"
                                            label="What gameplay are you interested in?"
                                            :items="gameplayList"></v-select>
                                    </v-col>
                                </v-row>

                                <v-row class="justify-center">
                                    <v-col cols="12">
                                        <v-combobox required v-model="timeZone" label="What timezone are you in?"
                                            :items="timeZones"></v-combobox>
                                    </v-col>
                                </v-row>

                                <v-row class="justify-center">
                                    <v-col cols="12">
                                        <v-btn color="primary" type="submit" @click="(e) => submit(e)"
                                            :disabled="!rsiHandleValid || checkingRSIHandle || timeZone == null">Submit</v-btn>
                                    </v-col>
                                </v-row>
                            </v-form>
                        </div>
                    </v-card>

                </v-col>
            </v-row>
        </v-responsive>
    </v-container>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useAppStore } from '@/stores/app'
import { useTheme } from 'vuetify'
import moment from 'moment-timezone'
import { useMembersStore } from '@/stores/members'
import { storeToRefs } from 'pinia'
import { Gameplay } from '@/stores/enums'

const appStore = useAppStore()

const theme = useTheme()
const logoColor = ref("blue")

const { me } = storeToRefs(appStore)

const foundBy = ref(null)
const recruitedBy = ref(null)
const rsiHandle = ref(null)
const checkingRSIHandle = ref(false)
const rsiHandleValid = ref(null)
const members = ref([])
const timeZones = moment.tz.names()
const timeZone = ref(me.value.timeZone)
const gameplayList = computed(() => {
    return Object.keys(Gameplay).map((key) => {
        return Gameplay[key]
    })
})
const chosenGameplay = ref([])

watch(chosenGameplay, () => {
    me.value.gameplay = chosenGameplay.value
})

watch(timeZone, () => {
    me.value.timeZone = timeZone.value
})

watch(foundBy, () => {
    me.value.foundBy = foundBy.value
    if (foundBy.value && foundBy.value == "recruited" && members.value.length == 0) {
        useMembersStore().getMembers().then((res) => {
            for (let i = 0; i < res.length; i++) {
                members.value.push({
                    title: res[i].name,
                    value: res[i].id
                })
            }
        })
    }
})

watch(recruitedBy, () => {
    if (recruitedBy.value) {
        me.value.recruitedBy = recruitedBy.value.value
    }
})

// debounce rsi handle input
let timeout = null
function debounce() {
    checkingRSIHandle.value = true
    clearTimeout(timeout)
    timeout = setTimeout(() => {
        appStore.checkRSIHandle(rsiHandle.value).then((result) => {
            rsiHandleValid.value = result
            checkingRSIHandle.value = false
        })
    }, 1000)
}

function submit(event) {
    event.preventDefault()

    appStore.updateSelf().then((res) => {
        if (res) {
            localStorage.setItem('onboarded', true)
            window.location.href = "/"
        }
    })
}

onMounted(() => {
    // if the theme is dark, set the logo to white
    if (theme.name.value == "dark") {
        logoColor.value = "white"
    }

    rsiHandle.value = me.value.name
    debounce()

    if (me.value.gameplay) {
        for (let i = 0; i < me.value.gameplay.length; i++) {
            chosenGameplay.value.push(Gameplay[me.value.gameplay[i]])
        }
    }

    if (me.value.foundBy) {
        foundBy.value = me.value.foundBy
    }
})

function isMobile() {
    return /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent)
}

</script>
<route lang="yaml">
meta:
    layout: plain
</route>
