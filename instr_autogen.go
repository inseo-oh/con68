// This file was automatically generated.
// Generated at 2025-04-07 14:33:00
package main

type instrLink struct {
    // Fields
    regY uint8
    
    imm16 uint16
}

type instrUnlk struct {
    // Fields
    regY uint8
    
}

type instrSwap struct {
    // Fields
    regY uint8
    
}

type instrMoveToUsp struct {
    // Fields
    regY uint8
    
}

type instrMoveFromUsp struct {
    // Fields
    regY uint8
    
}

type instrExtW struct {
    // Fields
    regY uint8
    
}

type instrExtL struct {
    // Fields
    regY uint8
    
}

type instrTrap struct {
    // Fields
    vector uint8
    
}

type instrTrapV struct {
    // Fields
    
}

type instrExgDReg struct {
    // Fields
    regX uint8
    regY uint8
    
}

type instrExgAReg struct {
    // Fields
    regX uint8
    regY uint8
    
}

type instrExgDAReg struct {
    // Fields
    regX uint8
    regY uint8
    
}

type instrIllegal struct {
    // Fields
    
}

type instrNop struct {
    // Fields
    
}

type instrRts struct {
    // Fields
    
}

type instrRtr struct {
    // Fields
    
}

type instrReset struct {
    // Fields
    
}

type instrRte struct {
    // Fields
    
}

//==========================================================================
// Decoder function
//==========================================================================
func (ctx *clientContext) instrDecode() (res instr, err error) {
    // instrLink
    func() {
        err = nil
        resTemp := instrLink{}
        if (ctx.decodingCtx.ir & 0xfff8) != 0x4e50 {
            err = excError{exc: excIllegalInstr}
            return
        }
        if v, ok := ctx.decodeFieldRegY(); !ok {
            err = excError{exc: excIllegalInstr}
            return
        }else {
            resTemp.regY = v
        }
        if !ctx.checkEaModes([]eamode{}, []eamode{}) {
            err = excError{exc: excIllegalInstr}
            return
        }
        if v, err := ctx.decodeXwordImm16(); err != nil {
            err = excError{exc: excIllegalInstr}
            return
        }else {
            resTemp.imm16 = v
        }
        if err = ctx.decodeEa(); err != nil {
            return
        }
        res = resTemp
    }()
    if excErr, isExcErr := err.(excError); !isExcErr || (isExcErr && (excErr.exc != excIllegalInstr)) {
        return
    }
    // instrUnlk
    func() {
        err = nil
        resTemp := instrUnlk{}
        if (ctx.decodingCtx.ir & 0xfff8) != 0x4e58 {
            err = excError{exc: excIllegalInstr}
            return
        }
        if v, ok := ctx.decodeFieldRegY(); !ok {
            err = excError{exc: excIllegalInstr}
            return
        }else {
            resTemp.regY = v
        }
        if !ctx.checkEaModes([]eamode{}, []eamode{}) {
            err = excError{exc: excIllegalInstr}
            return
        }
        if err = ctx.decodeEa(); err != nil {
            return
        }
        res = resTemp
    }()
    if excErr, isExcErr := err.(excError); !isExcErr || (isExcErr && (excErr.exc != excIllegalInstr)) {
        return
    }
    // instrSwap
    func() {
        err = nil
        resTemp := instrSwap{}
        if (ctx.decodingCtx.ir & 0xfff8) != 0x4840 {
            err = excError{exc: excIllegalInstr}
            return
        }
        if v, ok := ctx.decodeFieldRegY(); !ok {
            err = excError{exc: excIllegalInstr}
            return
        }else {
            resTemp.regY = v
        }
        if !ctx.checkEaModes([]eamode{}, []eamode{}) {
            err = excError{exc: excIllegalInstr}
            return
        }
        if err = ctx.decodeEa(); err != nil {
            return
        }
        res = resTemp
    }()
    if excErr, isExcErr := err.(excError); !isExcErr || (isExcErr && (excErr.exc != excIllegalInstr)) {
        return
    }
    // instrMoveToUsp
    func() {
        err = nil
        resTemp := instrMoveToUsp{}
        if (ctx.decodingCtx.ir & 0xfff8) != 0x4e60 {
            err = excError{exc: excIllegalInstr}
            return
        }
        if v, ok := ctx.decodeFieldRegY(); !ok {
            err = excError{exc: excIllegalInstr}
            return
        }else {
            resTemp.regY = v
        }
        if !ctx.checkEaModes([]eamode{}, []eamode{}) {
            err = excError{exc: excIllegalInstr}
            return
        }
        if err = ctx.decodeEa(); err != nil {
            return
        }
        res = resTemp
    }()
    if excErr, isExcErr := err.(excError); !isExcErr || (isExcErr && (excErr.exc != excIllegalInstr)) {
        return
    }
    // instrMoveFromUsp
    func() {
        err = nil
        resTemp := instrMoveFromUsp{}
        if (ctx.decodingCtx.ir & 0xfff8) != 0x4e68 {
            err = excError{exc: excIllegalInstr}
            return
        }
        if v, ok := ctx.decodeFieldRegY(); !ok {
            err = excError{exc: excIllegalInstr}
            return
        }else {
            resTemp.regY = v
        }
        if !ctx.checkEaModes([]eamode{}, []eamode{}) {
            err = excError{exc: excIllegalInstr}
            return
        }
        if err = ctx.decodeEa(); err != nil {
            return
        }
        res = resTemp
    }()
    if excErr, isExcErr := err.(excError); !isExcErr || (isExcErr && (excErr.exc != excIllegalInstr)) {
        return
    }
    // instrExtW
    func() {
        err = nil
        resTemp := instrExtW{}
        if (ctx.decodingCtx.ir & 0xfff8) != 0x4880 {
            err = excError{exc: excIllegalInstr}
            return
        }
        if v, ok := ctx.decodeFieldRegY(); !ok {
            err = excError{exc: excIllegalInstr}
            return
        }else {
            resTemp.regY = v
        }
        if !ctx.checkEaModes([]eamode{}, []eamode{}) {
            err = excError{exc: excIllegalInstr}
            return
        }
        if err = ctx.decodeEa(); err != nil {
            return
        }
        res = resTemp
    }()
    if excErr, isExcErr := err.(excError); !isExcErr || (isExcErr && (excErr.exc != excIllegalInstr)) {
        return
    }
    // instrExtL
    func() {
        err = nil
        resTemp := instrExtL{}
        if (ctx.decodingCtx.ir & 0xfff8) != 0x48c0 {
            err = excError{exc: excIllegalInstr}
            return
        }
        if v, ok := ctx.decodeFieldRegY(); !ok {
            err = excError{exc: excIllegalInstr}
            return
        }else {
            resTemp.regY = v
        }
        if !ctx.checkEaModes([]eamode{}, []eamode{}) {
            err = excError{exc: excIllegalInstr}
            return
        }
        if err = ctx.decodeEa(); err != nil {
            return
        }
        res = resTemp
    }()
    if excErr, isExcErr := err.(excError); !isExcErr || (isExcErr && (excErr.exc != excIllegalInstr)) {
        return
    }
    // instrTrap
    func() {
        err = nil
        resTemp := instrTrap{}
        if (ctx.decodingCtx.ir & 0xfff0) != 0x4e40 {
            err = excError{exc: excIllegalInstr}
            return
        }
        if v, ok := ctx.decodeFieldVector(); !ok {
            err = excError{exc: excIllegalInstr}
            return
        }else {
            resTemp.vector = v
        }
        if !ctx.checkEaModes([]eamode{}, []eamode{}) {
            err = excError{exc: excIllegalInstr}
            return
        }
        if err = ctx.decodeEa(); err != nil {
            return
        }
        res = resTemp
    }()
    if excErr, isExcErr := err.(excError); !isExcErr || (isExcErr && (excErr.exc != excIllegalInstr)) {
        return
    }
    // instrTrapV
    func() {
        err = nil
        resTemp := instrTrapV{}
        if (ctx.decodingCtx.ir & 0xffff) != 0x4e76 {
            err = excError{exc: excIllegalInstr}
            return
        }
        if !ctx.checkEaModes([]eamode{}, []eamode{}) {
            err = excError{exc: excIllegalInstr}
            return
        }
        if err = ctx.decodeEa(); err != nil {
            return
        }
        res = resTemp
    }()
    if excErr, isExcErr := err.(excError); !isExcErr || (isExcErr && (excErr.exc != excIllegalInstr)) {
        return
    }
    // instrExgDReg
    func() {
        err = nil
        resTemp := instrExgDReg{}
        if (ctx.decodingCtx.ir & 0xf1f8) != 0xc140 {
            err = excError{exc: excIllegalInstr}
            return
        }
        if v, ok := ctx.decodeFieldRegX(); !ok {
            err = excError{exc: excIllegalInstr}
            return
        }else {
            resTemp.regX = v
        }
        if v, ok := ctx.decodeFieldRegY(); !ok {
            err = excError{exc: excIllegalInstr}
            return
        }else {
            resTemp.regY = v
        }
        if !ctx.checkEaModes([]eamode{}, []eamode{}) {
            err = excError{exc: excIllegalInstr}
            return
        }
        if err = ctx.decodeEa(); err != nil {
            return
        }
        res = resTemp
    }()
    if excErr, isExcErr := err.(excError); !isExcErr || (isExcErr && (excErr.exc != excIllegalInstr)) {
        return
    }
    // instrExgAReg
    func() {
        err = nil
        resTemp := instrExgAReg{}
        if (ctx.decodingCtx.ir & 0xf1f8) != 0xc148 {
            err = excError{exc: excIllegalInstr}
            return
        }
        if v, ok := ctx.decodeFieldRegX(); !ok {
            err = excError{exc: excIllegalInstr}
            return
        }else {
            resTemp.regX = v
        }
        if v, ok := ctx.decodeFieldRegY(); !ok {
            err = excError{exc: excIllegalInstr}
            return
        }else {
            resTemp.regY = v
        }
        if !ctx.checkEaModes([]eamode{}, []eamode{}) {
            err = excError{exc: excIllegalInstr}
            return
        }
        if err = ctx.decodeEa(); err != nil {
            return
        }
        res = resTemp
    }()
    if excErr, isExcErr := err.(excError); !isExcErr || (isExcErr && (excErr.exc != excIllegalInstr)) {
        return
    }
    // instrExgDAReg
    func() {
        err = nil
        resTemp := instrExgDAReg{}
        if (ctx.decodingCtx.ir & 0xf1f8) != 0xc188 {
            err = excError{exc: excIllegalInstr}
            return
        }
        if v, ok := ctx.decodeFieldRegX(); !ok {
            err = excError{exc: excIllegalInstr}
            return
        }else {
            resTemp.regX = v
        }
        if v, ok := ctx.decodeFieldRegY(); !ok {
            err = excError{exc: excIllegalInstr}
            return
        }else {
            resTemp.regY = v
        }
        if !ctx.checkEaModes([]eamode{}, []eamode{}) {
            err = excError{exc: excIllegalInstr}
            return
        }
        if err = ctx.decodeEa(); err != nil {
            return
        }
        res = resTemp
    }()
    if excErr, isExcErr := err.(excError); !isExcErr || (isExcErr && (excErr.exc != excIllegalInstr)) {
        return
    }
    // instrIllegal
    func() {
        err = nil
        resTemp := instrIllegal{}
        if (ctx.decodingCtx.ir & 0xffff) != 0x4afc {
            err = excError{exc: excIllegalInstr}
            return
        }
        if !ctx.checkEaModes([]eamode{}, []eamode{}) {
            err = excError{exc: excIllegalInstr}
            return
        }
        if err = ctx.decodeEa(); err != nil {
            return
        }
        res = resTemp
    }()
    if excErr, isExcErr := err.(excError); !isExcErr || (isExcErr && (excErr.exc != excIllegalInstr)) {
        return
    }
    // instrNop
    func() {
        err = nil
        resTemp := instrNop{}
        if (ctx.decodingCtx.ir & 0xffff) != 0x4e71 {
            err = excError{exc: excIllegalInstr}
            return
        }
        if !ctx.checkEaModes([]eamode{}, []eamode{}) {
            err = excError{exc: excIllegalInstr}
            return
        }
        if err = ctx.decodeEa(); err != nil {
            return
        }
        res = resTemp
    }()
    if excErr, isExcErr := err.(excError); !isExcErr || (isExcErr && (excErr.exc != excIllegalInstr)) {
        return
    }
    // instrRts
    func() {
        err = nil
        resTemp := instrRts{}
        if (ctx.decodingCtx.ir & 0xffff) != 0x4e75 {
            err = excError{exc: excIllegalInstr}
            return
        }
        if !ctx.checkEaModes([]eamode{}, []eamode{}) {
            err = excError{exc: excIllegalInstr}
            return
        }
        if err = ctx.decodeEa(); err != nil {
            return
        }
        res = resTemp
    }()
    if excErr, isExcErr := err.(excError); !isExcErr || (isExcErr && (excErr.exc != excIllegalInstr)) {
        return
    }
    // instrRtr
    func() {
        err = nil
        resTemp := instrRtr{}
        if (ctx.decodingCtx.ir & 0xffff) != 0x4e77 {
            err = excError{exc: excIllegalInstr}
            return
        }
        if !ctx.checkEaModes([]eamode{}, []eamode{}) {
            err = excError{exc: excIllegalInstr}
            return
        }
        if err = ctx.decodeEa(); err != nil {
            return
        }
        res = resTemp
    }()
    if excErr, isExcErr := err.(excError); !isExcErr || (isExcErr && (excErr.exc != excIllegalInstr)) {
        return
    }
    // instrReset
    func() {
        err = nil
        resTemp := instrReset{}
        if (ctx.decodingCtx.ir & 0xffff) != 0x4e70 {
            err = excError{exc: excIllegalInstr}
            return
        }
        if !ctx.checkEaModes([]eamode{}, []eamode{}) {
            err = excError{exc: excIllegalInstr}
            return
        }
        if err = ctx.decodeEa(); err != nil {
            return
        }
        res = resTemp
    }()
    if excErr, isExcErr := err.(excError); !isExcErr || (isExcErr && (excErr.exc != excIllegalInstr)) {
        return
    }
    // instrRte
    func() {
        err = nil
        resTemp := instrRte{}
        if (ctx.decodingCtx.ir & 0xffff) != 0x4e73 {
            err = excError{exc: excIllegalInstr}
            return
        }
        if !ctx.checkEaModes([]eamode{}, []eamode{}) {
            err = excError{exc: excIllegalInstr}
            return
        }
        if err = ctx.decodeEa(); err != nil {
            return
        }
        res = resTemp
    }()
    if excErr, isExcErr := err.(excError); !isExcErr || (isExcErr && (excErr.exc != excIllegalInstr)) {
        return
    }
    return nil, excError{exc: excIllegalInstr}
}
