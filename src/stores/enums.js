const Ranks = {
    0: {
        id: 0,
        name: "None",
        short: "NIL",
        color: "bot",
    },
    1: {
        id: 1,
        name: "Admiral",
        short: "ADM",
        color: "admiral",
    },
    2: {
        id: 2,
        name: "Commander",
        short: "CMD",
        color: "commander",
    },
    3: {
        id: 3,
        name: "Lieutenant",
        short: "LT",
        color: "lieutenant",
    },
    4: {
        id: 4,
        name: "Specialist",
        short: "SPC",
        color: "specialist",
    },
    5: {
        id: 5,
        name: "Technician",
        short: "TEC",
        color: "technician",
    },
    6: {
        id: 6,
        name: "Member",
        short: "",
        color: "member",
    },
    7: {
        id: 7,
        name: "Recruit",
        short: "",
        color: "recruit",
    },
    8: {
        id: 8,
        name: "Guest",
        short: "",
        color: "guest",
    },
    99: {
        id: 99,
        name: "Ally",
        short: "",
        color: "ally",
    },
    100: {
        id: 100,
        name: "Pirate",
        short: "",
        color: "pirate",
    }
}

/*
    BountyHunting  GameplayTypes = "bounty_hunting"
    Engineering    GameplayTypes = "engineering"
    Exporation     GameplayTypes = "exporation"
    FpsCombat      GameplayTypes = "fps_combat"
    Hauling        GameplayTypes = "hauling"
    Medical        GameplayTypes = "medical"
    Mining         GameplayTypes = "mining"
    Reconnaissance GameplayTypes = "reconnaissance"
    Racing         GameplayTypes = "racing"
    Scrapping      GameplayTypes = "scrapping"
    ShipCrew       GameplayTypes = "ship_crew"
    ShipCombat     GameplayTypes = "ship_combat"
    Trading        GameplayTypes = "trading"
*/

const Gameplay = {
    "bounty_hunting": { title: "Bounty Hunting", value: "bounty_hunting" },
    "engineering": { title: "Engineering", value: "engineering" },
    "exporation": { title: "Exporation", value: "exporation" },
    "fps_combat": { title: "FPS Combat", value: "fps_combat" },
    "hauling": { title: "Hauling", value: "hauling" },
    "medical": { title: "Medical", value: "medical" },
    "mining": { title: "Mining", value: "mining" },
    "reconnaissance": { title: "Reconnaissance", value: "reconnaissance" },
    "racing": { title: "Racing", value: "racing" },
    "scrapping": { title: "Scrapping", value: "scrapping" },
    "ship_crew": { title: "Ship Crew", value: "ship_crew" },
    "ship_combat": { title: "Ship Combat", value: "ship_combat" },
    "trading": { title: "Trading", value: "trading" },
}

export { Ranks, Gameplay }
