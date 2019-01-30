package main

import (
    "math/big"
    "BlockChainTest/wasm/wiasm"
    "encoding/json"
    "math"
)

func RegisterUserByParams(p interface{}) interface{} {
    //wiasm.Log("exec RegisterUserByParams...")
    //wiasm.Log("params.(type) reflect: " + reflect.TypeOf(p).String())
    users := []*registerReturns{}
    switch p.(type) {
    case []interface{}:
        ps := p.([]interface{})
        for _, item := range ps {

            switch item.(type) {
            case string:
                param := &registerParams{}
                s := item.(string)
                err := json.Unmarshal([]byte(s), param)
                //wiasm.Log("RegisterUserByParams user" + s)
                user := generateOneUser(param)
                if err != nil {
                    wiasm.Log("RegisterUserByParams json.Unmarshal error" + err.Error())
                    return err. Error()
                }
                //wiasm.Log("RegisterUserByParams json.Unmarshal succeed!")
                users = append(users, user)
            }

        }
    }
    return users
}

func GainByParams(params interface{}) interface{} {
    wiasm.Log("GainByParams test : 1")
    par := &gainParams{}
    ret := &gainReturns{}

    switch params.(type) {
    case string:
       wiasm.Log("GainByParams test : 2")
       err := json.Unmarshal([]byte(params.(string)), par)
       if err != nil {
           ret.Error = "json unmarshall params error"
           return ret
       }
       for _, inc := range par.Increments {
           tracer, ok := par.Tracers[inc.RepID]
           if !ok || tracer == nil {
               continue
           }
           _, ok = par.Active[inc.RepID]
           if !ok {
               continue
           }
           if tracer.LastLiquidateDay == 0 {
               tracer.LastLiquidateDay = inc.Day
           }
           if !par.Active[inc.RepID] {
               par.Active[inc.RepID] = true
               tracer.FirstActiveDay = inc.Day
           }
           if _, ok := tracer.GainMemory[inc.Day]; ok {
               tracer.GainMemory[inc.Day] += inc.Gain
           } else {
               tracer.GainMemory[inc.Day] = inc.Gain
           }
       }

       ret.PlatformID = par.PlatformID
       //wiasm.Log("ret.PlatformID:" + string(par.PlatformID))
       ret.Tracers = par.Tracers
       ret.Active = par.Active
       ret.Error = ""

    }
    return ret
}

func AcceptModel(params interface{}) interface{} {
    par := &model{}

    switch params.(type) {
    case string:
        //time1 := time.Now()
        err := json.Unmarshal([]byte(params.(string)), par)
        if err != nil {
            return err.Error()
        }
        mod = par
        //wiasm.Log("AcceptModel Succeed! cost time: " + time.Now().Sub(time1).String())
    }
    return nil
}

func LiquidateByParams(params interface{}) interface{} {
    wiasm.Log("LiquidateByParams test :xc ")
    par := &liquidateParams{}
    ret := &liquidateReturns{}

    switch params.(type) {
    case string:
        err := json.Unmarshal([]byte(params.(string)), par)
        wiasm.Log("LiquidateByParams test :xc")
        if err != nil {
           ret.Error = "json unmarshal params error"
           return ret
        }
        for _, repID := range par.RepIDs {
            tracer, ok := par.Tracers[repID]
            if !ok || tracer == nil {
                continue
            }
            _, ok = par.Active[repID]
            if !ok {
                continue
            }
            for day := tracer.LastLiquidateDay; day < par.Until; day++ {
                gap := new(big.Float).SetInt64(int64(day - tracer.FirstActiveDay))
                var gain *big.Float
                var gValue int
                if value, ok := tracer.GainMemory[day]; ok {
                    gain = new(big.Float).SetInt64(int64(value))
                    gValue = value
                } else {
                    gain = new(big.Float)
                    gValue = 0
                }
                diff := new(big.Float).Sub(mod.Te, mod.T0)
                btRoot := new(big.Float).Sqrt(tracer.Bottom)
                btRoot.Mul(btRoot, new(big.Float).SetInt64(20))
                index0, _ := diff.Add(diff, btRoot).Float64()
                index := int(math.Ceil(index0))
                alpha2 := mod.Alpha2[int(index)]
                if alpha2 == nil {
                    continue
                }
                if gap.Cmp(mod.T0) < 0 {
                    tracer.Recent.Mul(tracer.Recent, mod.Alpha1)
                    tracer.Recent.Add(tracer.Recent, gain)
                } else {
                    delta := new(big.Float).Mul(mod.R0, new(big.Float).SetInt64(int64(tracer.GainHistory[0])))
                    tracer.Remote.Add(tracer.Remote, delta)
                    tracer.Remote.Mul(tracer.Remote, alpha2)
                }
                if gain.Cmp(mod.Require) >= 0 {
                    tracer.Aliveness.Add(tracer.Aliveness, one)
                } else {
                    tracer.Aliveness.Mul(tracer.Aliveness, mod.Epsilon)
                }

                ali := new(big.Float).Set(tracer.Aliveness)
                coef := new(big.Float).Mul(ali, ali)
                coef.Mul(coef, ali)
                coef.Sqrt(coef)
                coef.Sqrt(coef)
                coef.Mul(coef, mod.Beta)
                coef.Add(coef, one)

                rep := new(big.Float).Add(tracer.Recent, tracer.Remote)
                rep.Mul(rep, coef)
                if rep.Cmp(tracer.Bottom) < 0 {
                    rep.Set(tracer.Bottom)
                }
                bot := new(big.Float).Mul(rep, mod.BottomRate)
                if bot.Cmp(tracer.Bottom) > 0 {
                    tracer.Bottom.Set(bot)
                }
                tracer.Rep.Set(rep)

                if gap.Cmp(mod.T0) < 0 {
                    tracer.GainHistory = append(tracer.GainHistory, gValue)
                } else {
                    tracer.GainHistory = append(tracer.GainHistory[1:], gValue)
                }
                delete(tracer.GainMemory, day)
                if gValue == 0 && par.Active[repID] {
                    par.Active[repID] = false
                }
            }
            tracer.LastLiquidateDay = par.Until
        }
        ret.PlatformID = par.PlatformID
        ret.Tracers = par.Tracers
        ret.Active = par.Active
        ret.Error = ""
    }
    return ret
}