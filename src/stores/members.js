import { defineStore } from "pinia"
import { useConnectionStore } from "./connection"
import CommandResponse from "./classes"
import { Member } from "./classes"

export const useMembersStore = defineStore("members", () => {
    const connectionStore = useConnectionStore()

    // map of member id to member
    const members = ref({})

    function bindEvents() {
        connectionStore.$subscribe((mutation, state) => {
            if (state.isConnected) {
                connectionStore.socket.send("members|get")
        
                connectionStore.socket.addEventListener("message", (event) => {
                    const commandResponse = new CommandResponse(event.data)

                    if (commandResponse.thing === "members") {
                        commandResponse.result.forEach((memberJson) => {
                            members.value[memberJson.id] = new Member(memberJson)
                        })
                    }
                })
            }
        })

        // sync the list of items upon connection
        //   socket.on("connect", () => {
        //     socket.emit("item:list", (res) => {
        //       this.items = res.data
        //     });
        //   });

        //   // update the store when an item was created
        //   socket.on("item:created", (item) => {
        //     this.items.push(item)
        //   });
    }

    function createItem(label) {
        // const item = {
        //     id: Date.now(), // temporary ID for v-for key
        //     label
        // };
        // items.push(item)

        //   socket.emit("item:create", { label }, (res) => {
        //     item.id = res.data
        //   })
    }

    return {
        members,
        bindEvents,
        createItem,
    }
})
