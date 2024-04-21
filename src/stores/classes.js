import { Ranks } from "./enums"

export const CommandResponse = class CommandReponse {
    constructor(j) {
        // convert to json
        j = JSON.parse(j)

        this.thing = j.thing
        this.action = j.action
        this.result = j.result
        this.error = j.error
    }
}

export const Member = class Member {
    constructor(memberJson) {
        this.id = memberJson.id
        this.name = memberJson.name
        this.rank = Ranks[memberJson.rank]
        this.eventsAttended = memberJson.events
        this.validated = memberJson.validated
        this.avatar = memberJson.avatar
    }

    /** @type {string} */
    static id
    /** @type {string} */
    static name
    /** @type {Object} */
    static rank
    /** @type {number} */
    static eventsAttended
    /** @type {bool} */
    static validated

    get officer() {
        return this.isOfficer()
    }

    isOfficer() {
        return this.rank.id <= 3
    }
}
