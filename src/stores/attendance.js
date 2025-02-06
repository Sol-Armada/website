import { defineStore } from "pinia"
import { useConnectionStore } from "./connection"
import { Attendance } from "./classes"
import { useErrorStore } from "./error"

export const useAttendanceStore = defineStore("attendance", {
    state: () => ({
        attendance: []
    }),
    actions: {
        async getAttendanceRecords(page) {
            return new Promise((resolve) => {
                const errorStore = useErrorStore()
                const connectionStore = useConnectionStore()
                connectionStore.addListener('attendance', 'list').then((commandResponse) => {
                    // handle errors
                    if (commandResponse.error) {
                        errorStore.$patch({ error: commandResponse.error, show: true })
                        return
                    }

                    if (!commandResponse.result) {
                        resolve([])
                        return
                    }

                    this.attendance.push(...commandResponse.result.map((a) => new Attendance(a)))

                    resolve(commandResponse.result.map((a) => new Attendance(a)))
                }).catch((error) => {
                    console.log(error)
                })

                setTimeout(() => {
                    connectionStore.send('attendance', 'list', page)
                }, 500)
            })
        },

        async getAttendanceCount(id) {
            return new Promise((resolve) => {
                const errorStore = useErrorStore()
                const connectionStore = useConnectionStore()
                connectionStore.addListener('attendance', 'count').then((commandResponse) => {
                    // handle errors
                    if (commandResponse.error) {
                        errorStore.$patch({ error: commandResponse.error, show: true })
                        return
                    }

                    resolve(commandResponse.result)
                }).catch((error) => {
                    console.log(error)
                })

                setTimeout(() => {
                    connectionStore.send('attendance', 'count', id)
                }, 500)
            })
        },

        async getMemberAttendanceRecords(id) {
            return new Promise((resolve) => {
                const errorStore = useErrorStore()
                const connectionStore = useConnectionStore()
                connectionStore.addListener('attendance', 'records').then((commandResponse) => {
                    // handle errors
                    if (commandResponse.error) {
                        errorStore.$patch({ error: commandResponse.error, show: true })
                        return
                    }

                    if (!commandResponse.result) {
                        resolve([])
                        return
                    }

                    resolve(commandResponse.result.map((a) => new Attendance(a)))
                }).catch((error) => {
                    console.log(error)
                })

                setTimeout(() => {
                    connectionStore.send('attendance', 'records', id)
                }, 500)
            })
        }
    }
})
