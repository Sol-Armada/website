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
        this.eventsAttended = memberJson.legacy_events
        this.validated = memberJson.validated
        this.avatar = memberJson.avatar

        this.age = memberJson.age
        this.playTime = memberJson.playtime
        this.validated = memberJson.validated
        this.gameplay = memberJson.gameplay
        this.onboarded_at = memberJson.onboarded_at ? new Date(memberJson.onboarded_at) : null
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
    /** @type {string} */
    static avatar
    /** @type {Date} */
    static onboarded_at

    get officer() {
        return this.isOfficer()
    }

    isOfficer() {
        return this.rank.id <= 3
    }
}

export const Attendance = class Attendance {
    constructor(attendanceJson) {
        this.id = attendanceJson.id
        this.name = attendanceJson.name
        this.dateCreated = new Date(attendanceJson.dateCreated)
        this.members = []
        if (attendanceJson.members) {
            for (let i = 0; i < attendanceJson.members.length; i++) {
                this.members.push(new Member(attendanceJson.members[i]))
            }
        }
        this.membersWithIssues = []
        if (attendanceJson.issues) {
            for (let i = 0; i < attendanceJson.issues.length; i++) {
                this.membersWithIssues.push(new Member(attendanceJson.issues[i]))
            }
        }
        this.recorded = attendanceJson.recorded
        this.submittedBy = new Member(attendanceJson.submitted_by)
    }

    /** @type {string} */
    static id
    /** @type {string} */
    static name
    /** @type {Date} */
    static dateCreated
    /** @type {Member[]} */
    static members
    /** @type {Member[]} */
    static membersWithIssues
    /** @type {bool} */
    static recorded
    /** @type {Member} */
    static submittedBy

    get numberOfMembers() {
        return this.numberOfMembers()
    }

    numberOfMembers() {
        return this.members.length
    }
}
