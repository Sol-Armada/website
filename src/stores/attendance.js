import { defineStore } from "pinia"
import { useConnectionStore } from "./connection"
import { Attendance } from "./classes"
import { useErrorStore } from "./error"

export const useAttendanceStore = defineStore("attendance", {
    actions: {
        async getAttendanceRecords(page) {
            return new Promise((resolve) => {
                const errorStore = useErrorStore()
                const connectionStore = useConnectionStore()
                connectionStore.addListener('attendance', 'list', (commandResponse) => {
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
                })

                setTimeout(() => {
                    connectionStore.send('attendance', 'list', page)
                }, 500)
            })
        }
    }
})
