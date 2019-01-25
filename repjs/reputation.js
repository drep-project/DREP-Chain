require("repjs/bignumber.js")

function DbPersistent() { }

DbPersistent.prototype.getTracer = function (platformID, repID) {
    var persistentObject = db.getTracer(platformID, repID)
    if (persistentObject) {
        return {
            Rep: new BigNumber(persistentObject.Rep),
            Aliveness: new BigNumber(persistentObject.Aliveness),
            Recent: new BigNumber(persistentObject.Recent),
            Remote: new BigNumber(persistentObject.Remote),
            Bottom: new BigNumber(persistentObject.Bottom),
            GainHistory: persistentObject.GainHistory,
            GainMemory: persistentObject.GainMemory,
            FirstActiveDay: persistentObject.FirstActiveDay,
            LastLiquidateDay: persistentObject.LastLiquidateDay
        }
    }
}

DbPersistent.prototype.putTracer = function (platformID, repID, jValue) {
    var persistentObject = {
        Rep: jValue.Rep.toString(),
        Aliveness: jValue.Aliveness.toString(),
        Recent: jValue.Recent.toString(),
        Remote: jValue.Remote.toString(),
        Bottom: jValue.Bottom.toString(),
        GainHistory: jValue.GainHistory,
        GainMemory: jValue.GainMemory,
        FirstActiveDay: jValue.FirstActiveDay,
        LastLiquidateDay: jValue.LastLiquidateDay
    }
    db.putTracer(platformID, repID, persistentObject)
}

DbPersistent.prototype.isActive = function (platformID, repID) {
    return db.isActive(platformID, repID)
}

DbPersistent.prototype.setActive = function (platformID, repID, active) {
    db.setActive(platformID, repID, active)
}

DbPersistent.prototype.getGroup = function (platformID, groupID) {
    return db.getGroup(platformID, groupID)
}

DbPersistent.prototype.putGroup = function (platformID, groupID, jValue) {
    db.putGroup(platformID, groupID, jValue)
}

var dbPersistent = new DbPersistent()

var rep_configs = require("repjs/data/rep_configs.json")
var rep_floats = require('repjs/data/rep_floats.json')

var r0 = new BigNumber(0)
var t0 = new BigNumber(0)
var te = new BigNumber(0)
var alpha1 = new BigNumber(0)
var alpha2 = {}

var decaymode = rep_configs.decaymode
if (decaymode === "faster") {
    r0 = new BigNumber(rep_floats.r1)
    t0 = new BigNumber(rep_floats.t1)
    te = new BigNumber(rep_floats.e1)
    alpha1 = new BigNumber(rep_floats.alpha1_1)
    for (var i = 0; i < 2000; i++) {
        alpha2[i] = rep_floats["alpha2_1_" + i]
    }
}
if (decaymode === "fast") {
    r0 = new BigNumber(rep_floats.r2)
    t0 = new BigNumber(rep_floats.t2)
    te = new BigNumber(rep_floats.e2)
    alpha1 = new BigNumber(rep_floats.alpha1_2)
    for (var i = 0; i < 2000; i++) {
        alpha2[i] = rep_floats["alpha2_2_" + i]
    }
}
if (decaymode === "medium") {
    r0 = new BigNumber(rep_floats.r3)
    t0 = new BigNumber(rep_floats.t3)
    te = new BigNumber(rep_floats.e3)
    alpha1 = new BigNumber(rep_floats.alpha1_3)
    for (var i = 0; i < 2000; i++) {
        alpha2[i] = rep_floats["alpha2_3_" + i]
    }
}
if (decaymode === "slow") {
    r0 = new BigNumber(rep_floats.r4)
    t0 = new BigNumber(rep_floats.t4)
    te = new BigNumber(rep_floats.e4)
    alpha1 = new BigNumber(rep_floats.alpha1_4)
    for (var i = 0; i < 2000; i++) {
        alpha2[i] = rep_floats["alpha2_4_" + i]
    }
}
if (decaymode === "slower") {
    r0 = new BigNumber(rep_floats.r5)
    t0 = new BigNumber(rep_floats.t5)
    te = new BigNumber(rep_floats.e5)
    alpha1 = new BigNumber(rep_floats.alpha1_5)
    for (var i = 0; i < 2000; i++) {
        alpha2[i] = rep_floats["alpha2_5_" + i]
    }
}

var bottomRate = new BigNumber(rep_configs.bottomrate)
var beta = new BigNumber(rep_configs.beta)
var epsilon = new BigNumber(rep_configs.epsilon)
var threshold = new BigNumber(rep_configs.threshold)
var one = new BigNumber(1)

function getProfile(platformID, uid) {
    var b1 = crypto.hash256(utils.str2Bytes(uid))
    var b2 = crypto.hash256(utils.str2Bytes(platformID))
    var b = crypto.hash256(b1, b2)
    var repID = utils.bytes2Hex(b)
    var groupID = utils.allocateGroupID(repID)
    return {RepID: repID, GroupID: groupID}
}

function registerUser(platformID, repID, groupID) {
    var group = dbPersistent.getGroup(platformID, groupID)
    group.push(repID)
    var tracer = {
        Rep: new BigNumber(0),
        Aliveness: new BigNumber(0),
        Recent: new BigNumber(0),
        Remote: new BigNumber(0),
        Bottom: new BigNumber(0),
        GainHistory: [],
        GainMemory: {},
        FirstActiveDay: 0,
        LastLiquidateDay: 0,
    }
    dbPersistent.putGroup(platformID, groupID, group)
    dbPersistent.setActive(platformID, repID, false)
    dbPersistent.putTracer(platformID, repID, tracer)
}

function registerUsers(platformID, uids) {
    var ret = []
    for (var i = 0; i < uids.length; i++ ) {
        var uid = uids[i]
        var b1 = crypto.hash256(utils.str2Bytes(uid))
        var b2 = crypto.hash256(utils.str2Bytes(platformID))
        var b = crypto.hash256(b1, b2)
        var repID = utils.bytes2Hex(b)
        var groupID = utils.allocateGroupID(repID)
        var group = dbPersistent.getGroup(platformID, groupID)
        group.push(repID)
        var tracer = {
            Rep: new BigNumber(0),
            Aliveness: new BigNumber(0),
            Recent: new BigNumber(0),
            Remote: new BigNumber(0),
            Bottom: new BigNumber(0),
            GainHistory: [],
            GainMemory: {},
            FirstActiveDay: 0,
            LastLiquidateDay: -1,
        }
        dbPersistent.putGroup(platformID, groupID, group)
        dbPersistent.setActive(platformID, repID, false)
        dbPersistent.putTracer(platformID, repID, tracer)
        ret.push({repID: repID, groupID: groupID})
    }
    return ret
}

function addGain(platformID, increments) {
    for (var i = 0; i < increments.length; i++) {
        var inc = increments[i]
        var repID = inc.RepID
        var gain = inc.Gain
        var day = inc.Day
        var tracer = dbPersistent.getTracer(platformID, repID)
        if (!tracer) {
            return
        }
        var active = dbPersistent.isActive(platformID, repID)
        if (tracer.LastLiquidateDay === -1) {
            tracer.LastLiquidateDay = day
        }
        if (!active) {
            dbPersistent.setActive(platformID, repID, true)
        }
        if (tracer.GainMemory[day]) {
            tracer.GainMemory[day] += gain
        } else {
            tracer.GainMemory[day] = gain
        }
        dbPersistent.putTracer(platformID, repID, tracer)
    }
}

function liquidateRep(platformID, repIDs, until) {
    for (var i = 0; i < repIDs.length; i++) {
        var repID = repIDs[i]
        var tracer = dbPersistent.getTracer(platformID, repID)
        if (!tracer) {
            continue
        }
        var active = dbPersistent.isActive(platformID, repID)
        for (var day = tracer.LastLiquidateDay; day < until; day++) {
            var gap = new BigNumber(day - tracer.FirstActiveDay)
            var gain = new BigNumber(0)
            var gValue = 0
            if (tracer.GainMemory[day]) {
                gain = new BigNumber(tracer.GainMemory[day])
                gValue = tracer.GainMemory[day]
            }

            var sbt = new BigNumber(0).plus(tracer.Bottom).sqrt()
            sbt = sbt.mul(new BigNumber(20))
            var diff = new BigNumber(0).plus(te).minus(t0)
            diff = diff.plus(sbt)
            var index = diff.ceil()
            var aph2 = alpha2[index]
            if (!aph2) {
                throw new Error("error occurred in calculating aph2")
            }

            if (gap.lessThan(t0)) {
                tracer.Recent = tracer.Recent.mul(alpha1).plus(gain)
            } else {
                var delta = new BigNumber(0).plus(r0).mul(new BigNumber(tracer.GainHistory[0]))
                tracer.Remote = tracer.Remote.plus(delta).mul(aph2)
            }

            if (gain.greaterThanOrEqualTo(threshold)) {
                tracer.Aliveness = tracer.Aliveness.plus(one)
            } else {
                tracer.Aliveness = tracer.Aliveness.mul(epsilon)
            }

            var ali = new BigNumber(0).plus(tracer.Aliveness)
            var coef = new BigNumber(0).plus(ali).mul(ali)
            coef = coef.mul(ali).sqrt().sqrt().mul(beta).plus(one)

            var rep = new BigNumber(0).plus(tracer.Recent).plus(tracer.Remote).mul(coef)
            if (rep.lessThan(tracer.Bottom)) {
                rep = new BigNumber(0).plus(tracer.Bottom)
                if (active) {
                    active = true
                }
            }

            var bot = new BigNumber(0).plus(rep).mul(bottomRate)
            if (bot.greaterThan(tracer.Bottom)) {
                tracer.Bottom = new BigNumber(0).plus(bot)
            }

            tracer.Rep = new BigNumber(0).plus(rep)

            if (gap.lessThan(t0)) {
                tracer.GainHistory.push(gValue)
            } else {
                tracer.GainHistory.shift()
                tracer.GainHistory.push(gValue)
            }
            delete tracer.GainMemory[day]
        }
        tracer.LastLiquidateDay = until
        dbPersistent.setActive(platformID, repID, active)
        dbPersistent.putTracer(platformID, repID, tracer)
        console.log()
        console.log("tracer: ", tracer.Rep, "  ", tracer.Recent, "  ", tracer.Aliveness)
        var keys1 = Object.keys(tracer.GainMemory)
        for(var j = 0; j < keys1.length; j++) {
            console.log(keys1[j] + ": " + tracer.GainMemory[keys1[j]])
        }
        console.log()
    }
}

function liquidateRepByGroup(platformID, groupID, until) {
    var group = dbPersistent.getGroup(platformID, groupID)
    liquidateRep(platformID, group, until)
}