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
        if (typeof memberJson.rank == 'object') {
            this.rank = Ranks[memberJson.rank.id]
        } else {
            this.rank = Ranks[memberJson.rank]
        }
        this.eventsAttended = memberJson.eventsAttended
        this.validated = memberJson.validated
        this.avatar = memberJson.avatar
        this.isGuest = memberJson.is_guest
        this.isBot = memberJson.is_bot
        this.isAlly = memberJson.is_ally
        this.isAffiliate = memberJson.is_affiliate

        this.timeZone = memberJson.time_zone
        this.foundBy = memberJson.found_by
        if (memberJson.recruiter) {
            this.recruitedBy = new Member(memberJson.recruiter)
        }
        this.age = memberJson.age
        this.playTime = memberJson.playtime
        this.validated = memberJson.validated
        this.gameplay = memberJson.gameplay
        this.onboardedAt = memberJson.onboarded_at ? new Date(memberJson.onboarded_at) : null
        this.other = memberJson.other
    }

    get onboarded() {
        return this.onboardedAt !== null
    }

    get isOfficer() {
        return this.rank.id <= 3 && this.rank.id != 0
    }

    get isMember() {
        return !this.isGuest && !this.isBot && !this.isAlly && !this.isAffiliate
    }

    get affiliation() {
        if (this.isAlly) {
            return "Ally"
        } else if (this.isAffiliate) {
            return "Affiliate"
        } else {
            return "Guest"
        }
    }
}

export const Attendance = class Attendance {
    constructor(attendanceJson) {
        this.id = attendanceJson.id
        this.name = attendanceJson.name
        this.dateCreated = new Date(attendanceJson.date_created)
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

    get numberOfMembers() {
        return this.members.length
    }

    get createdDate() {
        const utcDate = new Date(this.dateCreated)
        const options = {
            year: 'numeric',
            month: 'long',
            day: 'numeric',
        }
        return utcDate.toLocaleString(undefined, options)
    }
}

export const Token = class Token {
    constructor(tokenJson) {
        this.id = tokenJson.id
        this.memberId = tokenJson.member_id
        this.reason = tokenJson.reason
        this.comment = tokenJson.comment
        this.amount = tokenJson.amount
        this.attendanceId = tokenJson.attendance_id

        this.createdAt = new Date(tokenJson.created_at)
    }

    get createdDate() {
        const utcDate = new Date(this.createdAt)
        const options = {
            year: 'numeric',
            month: 'long',
            day: 'numeric',
        }
        return utcDate.toLocaleString(undefined, options)
    }
}
