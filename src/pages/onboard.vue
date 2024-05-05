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
                                        <v-combobox label="How did you find us?"
                                            :items="[{ title: 'RSI Website/Spectrum', value: 'rsi_website' }, { title: 'Recruited by a member', value: 'recruited' }, { title: 'Other', value: 'other' }]"
                                            v-model="foundBy"></v-combobox>
                                    </v-col>
                                </v-row>

                                <v-row v-if="foundBy && foundBy.value == 'recruited'" class="justify-center">
                                    <v-col cols="12">
                                        <v-combobox label="Who recruited you?"
                                            :items="[{ title: 'Person A', value: '1234556' }]"></v-combobox>
                                    </v-col>
                                </v-row>

                                <v-row v-if="foundBy && foundBy.value == 'other'" class="justify-center">
                                    <v-col cols="12">
                                        <v-textarea label="How did you find us?"></v-textarea>
                                    </v-col>
                                </v-row>

                                <v-row class="justify-center">
                                    <v-col cols="12" md="6">
                                        <h5>Years playing Star Citizen (Round up)</h5>
                                        <v-number-input :reverse="false" controlVariant="split" label=""
                                            :hideInput="false" :inset="false" v-model="yearsPlaying" min=0
                                            max=10></v-number-input>
                                    </v-col>
                                    <v-col cols="12" md="6">
                                        <h5>How old are you?</h5>
                                        <v-number-input :reverse="false" controlVariant="split" label=""
                                            :hideInput="false" :inset="false" v-model="age" min=16
                                            max=99></v-number-input>
                                    </v-col>
                                </v-row>

                                <v-row class="justify-center">
                                    <v-col cols="12">
                                        <v-select chips multiple hide-selected v-model="chosenGameplay"
                                            label="What gameplay are you interested in?"
                                            :items="gameplayList"></v-select>
                                    </v-col>
                                </v-row>

                                <v-row class="justify-center">
                                    <v-col cols="12">
                                        <v-btn color="primary" type="submit" @click="(e) => submit(e)"
                                            :disabled="!rsiHandleValid || checkingRSIHandle">Submit</v-btn>
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
import { onMounted, ref } from 'vue'
import { useAppStore } from '@/stores/app'
import { useTheme } from 'vuetify'

const appStore = useAppStore()

const theme = useTheme()
const logoColor = ref("blue")

const foundBy = ref(null)
const rsiHandle = ref(null)
const checkingRSIHandle = ref(false)
const rsiHandleValid = ref(null)
const yearsPlaying = ref(0)
const age = ref(16)
const gameplayList = [
    { title: "Mining", value: "mining" },
    { title: "Salvage", value: "salvage" },
    { title: "Shipping", value: "shipping" },
    { title: "Bounty Hunting", value: "bounty hunting" },
    { title: "Racing", value: "racing" },
    { title: "FPS Combat", value: "fps combat" },
    { title: "Ship Combat", value: "ship combat" },
    { title: "Engineering", value: "engineering" },
    { title: "Medical", value: "medical" },
    { title: "Exploration", value: "exploration" },
]
const chosenGameplay = ref([])

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

    appStore.me.age = age.value
    appStore.me.name = rsiHandle.value
    appStore.me.playtime = yearsPlaying.value
    const gamePlay = []
    for (let i = 0; i < chosenGameplay.value.length; i++) {
        gamePlay.push(chosenGameplay.value[i].value)
    }
    appStore.me.gameplay = gamePlay

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
})

</script>
<route lang="yaml">
meta:
  layout: plain
</route>
