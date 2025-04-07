// This file was automatically generated.
// Generated at 2025-04-07 17:54:23
package main

type instrMoveB struct {
    instrPc uint32
    
    ea1 *ea
    ea2 *ea
    
}

type instrBra struct {
    instrPc uint32
    
    
    branchOff uint32
}

type instrBsr struct {
    instrPc uint32
    
    
    branchOff uint32
}

type instrBcc struct {
    instrPc uint32
    
    cond cond
    
    branchOff uint32
}

type instrDbcc struct {
    instrPc uint32
    
    cond cond
    regY uint8
    
    imm16 uint16
}

type instrLea struct {
    instrPc uint32
    
    regX uint8
    ea1 *ea
    
}

type instrPea struct {
    instrPc uint32
    
    ea1 *ea
    
}

type instrJmp struct {
    instrPc uint32
    
    ea1 *ea
    
}

type instrJsr struct {
    instrPc uint32
    
    ea1 *ea
    
}

type instrLink struct {
    instrPc uint32
    
    regY uint8
    
    imm16 uint16
}

type instrUnlk struct {
    instrPc uint32
    
    regY uint8
    
}

type instrSwap struct {
    instrPc uint32
    
    regY uint8
    
}

type instrMoveToUsp struct {
    instrPc uint32
    
    regY uint8
    
}

type instrMoveFromUsp struct {
    instrPc uint32
    
    regY uint8
    
}

type instrExtW struct {
    instrPc uint32
    
    regY uint8
    
}

type instrExtL struct {
    instrPc uint32
    
    regY uint8
    
}

type instrTrap struct {
    instrPc uint32
    
    vector uint8
    
}

type instrTrapV struct {
    instrPc uint32
    
    
}

type instrExgDReg struct {
    instrPc uint32
    
    regX uint8
    regY uint8
    
}

type instrExgAReg struct {
    instrPc uint32
    
    regX uint8
    regY uint8
    
}

type instrExgDAReg struct {
    instrPc uint32
    
    regX uint8
    regY uint8
    
}

type instrIllegal struct {
    instrPc uint32
    
    
}

type instrNop struct {
    instrPc uint32
    
    
}

type instrRts struct {
    instrPc uint32
    
    
}

type instrRtr struct {
    instrPc uint32
    
    
}

type instrReset struct {
    instrPc uint32
    
    
}

type instrRte struct {
    instrPc uint32
    
    
}

//==========================================================================
// Decoder function
//==========================================================================
func (ctx *clientContext) instrDecode() (res instr, err error) {
    // instrMoveB
    func() {
        err = nil
        resTemp := instrMoveB{}
        resTemp.instrPc = ctx.pc - 2
        if (ctx.decodingCtx.ir & 0xf000) != 0x1000 {
            err = excError{exc: excIllegalInstr}
            return
        }
        if v, ok := ctx.decodeFieldEa1(); !ok {
            err = excError{exc: excIllegalInstr}
            return
        }else {
            resTemp.ea1 = v
        }
        if v, ok := ctx.decodeFieldEa2(); !ok {
            err = excError{exc: excIllegalInstr}
            return
        }else {
            resTemp.ea2 = v
        }
        if !ctx.checkEaModes([]eamode{eamodeDreg, eamodeAreg, eamodeAregInd, eamodeAregIndPostinc, eamodeAregIndPredec, eamodeAregIndDisp, eamodeAregIndIndex, eamodeAbsW, eamodeAbsL, eamodePcIndDisp, eamodePcIndIndex, eamodeImm}, []eamode{eamodeDreg, eamodeAregInd, eamodeAregIndPostinc, eamodeAregIndPredec, eamodeAregIndDisp, eamodeAregIndIndex, eamodeAbsW, eamodeAbsL}) {
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
    // instrBra
    func() {
        err = nil
        resTemp := instrBra{}
        resTemp.instrPc = ctx.pc - 2
        if (ctx.decodingCtx.ir & 0xff00) != 0x6000 {
            err = excError{exc: excIllegalInstr}
            return
        }
        if !ctx.checkEaModes([]eamode{}, []eamode{}) {
            err = excError{exc: excIllegalInstr}
            return
        }
        if v, err := ctx.decodeXwordBranchOff(); err != nil {
            err = excError{exc: excIllegalInstr}
            return
        }else {
            resTemp.branchOff = v
        }
        if err = ctx.decodeEa(); err != nil {
            return
        }
        res = resTemp
    }()
    if excErr, isExcErr := err.(excError); !isExcErr || (isExcErr && (excErr.exc != excIllegalInstr)) {
        return
    }
    // instrBsr
    func() {
        err = nil
        resTemp := instrBsr{}
        resTemp.instrPc = ctx.pc - 2
        if (ctx.decodingCtx.ir & 0xff00) != 0x6100 {
            err = excError{exc: excIllegalInstr}
            return
        }
        if !ctx.checkEaModes([]eamode{}, []eamode{}) {
            err = excError{exc: excIllegalInstr}
            return
        }
        if v, err := ctx.decodeXwordBranchOff(); err != nil {
            err = excError{exc: excIllegalInstr}
            return
        }else {
            resTemp.branchOff = v
        }
        if err = ctx.decodeEa(); err != nil {
            return
        }
        res = resTemp
    }()
    if excErr, isExcErr := err.(excError); !isExcErr || (isExcErr && (excErr.exc != excIllegalInstr)) {
        return
    }
    // instrBcc
    func() {
        err = nil
        resTemp := instrBcc{}
        resTemp.instrPc = ctx.pc - 2
        if (ctx.decodingCtx.ir & 0xf000) != 0x6000 {
            err = excError{exc: excIllegalInstr}
            return
        }
        if v, ok := ctx.decodeFieldCond(); !ok {
            err = excError{exc: excIllegalInstr}
            return
        }else {
            resTemp.cond = v
        }
        if !ctx.checkEaModes([]eamode{}, []eamode{}) {
            err = excError{exc: excIllegalInstr}
            return
        }
        if v, err := ctx.decodeXwordBranchOff(); err != nil {
            err = excError{exc: excIllegalInstr}
            return
        }else {
            resTemp.branchOff = v
        }
        if err = ctx.decodeEa(); err != nil {
            return
        }
        res = resTemp
    }()
    if excErr, isExcErr := err.(excError); !isExcErr || (isExcErr && (excErr.exc != excIllegalInstr)) {
        return
    }
    // instrDbcc
    func() {
        err = nil
        resTemp := instrDbcc{}
        resTemp.instrPc = ctx.pc - 2
        if (ctx.decodingCtx.ir & 0xf0f8) != 0x50c8 {
            err = excError{exc: excIllegalInstr}
            return
        }
        if v, ok := ctx.decodeFieldCond(); !ok {
            err = excError{exc: excIllegalInstr}
            return
        }else {
            resTemp.cond = v
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
    // instrLea
    func() {
        err = nil
        resTemp := instrLea{}
        resTemp.instrPc = ctx.pc - 2
        if (ctx.decodingCtx.ir & 0xf1c0) != 0x41c0 {
            err = excError{exc: excIllegalInstr}
            return
        }
        if v, ok := ctx.decodeFieldRegX(); !ok {
            err = excError{exc: excIllegalInstr}
            return
        }else {
            resTemp.regX = v
        }
        if v, ok := ctx.decodeFieldEa1(); !ok {
            err = excError{exc: excIllegalInstr}
            return
        }else {
            resTemp.ea1 = v
        }
        if !ctx.checkEaModes([]eamode{eamodeAregInd, eamodeAregIndDisp, eamodeAregIndIndex, eamodeAbsW, eamodeAbsL, eamodePcIndDisp, eamodePcIndIndex}, []eamode{}) {
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
    // instrPea
    func() {
        err = nil
        resTemp := instrPea{}
        resTemp.instrPc = ctx.pc - 2
        if (ctx.decodingCtx.ir & 0xffc0) != 0x4840 {
            err = excError{exc: excIllegalInstr}
            return
        }
        if v, ok := ctx.decodeFieldEa1(); !ok {
            err = excError{exc: excIllegalInstr}
            return
        }else {
            resTemp.ea1 = v
        }
        if !ctx.checkEaModes([]eamode{eamodeAregInd, eamodeAregIndDisp, eamodeAregIndIndex, eamodeAbsW, eamodeAbsL, eamodePcIndDisp, eamodePcIndIndex}, []eamode{}) {
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
    // instrJmp
    func() {
        err = nil
        resTemp := instrJmp{}
        resTemp.instrPc = ctx.pc - 2
        if (ctx.decodingCtx.ir & 0xffc0) != 0x4ec0 {
            err = excError{exc: excIllegalInstr}
            return
        }
        if v, ok := ctx.decodeFieldEa1(); !ok {
            err = excError{exc: excIllegalInstr}
            return
        }else {
            resTemp.ea1 = v
        }
        if !ctx.checkEaModes([]eamode{eamodeAregInd, eamodeAregIndDisp, eamodeAregIndIndex, eamodeAbsW, eamodeAbsL, eamodePcIndDisp, eamodePcIndIndex}, []eamode{}) {
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
    // instrJsr
    func() {
        err = nil
        resTemp := instrJsr{}
        resTemp.instrPc = ctx.pc - 2
        if (ctx.decodingCtx.ir & 0xffc0) != 0x4e80 {
            err = excError{exc: excIllegalInstr}
            return
        }
        if v, ok := ctx.decodeFieldEa1(); !ok {
            err = excError{exc: excIllegalInstr}
            return
        }else {
            resTemp.ea1 = v
        }
        if !ctx.checkEaModes([]eamode{eamodeAregInd, eamodeAregIndDisp, eamodeAregIndIndex, eamodeAbsW, eamodeAbsL, eamodePcIndDisp, eamodePcIndIndex}, []eamode{}) {
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
    // instrLink
    func() {
        err = nil
        resTemp := instrLink{}
        resTemp.instrPc = ctx.pc - 2
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
        resTemp.instrPc = ctx.pc - 2
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
        resTemp.instrPc = ctx.pc - 2
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
        resTemp.instrPc = ctx.pc - 2
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
        resTemp.instrPc = ctx.pc - 2
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
        resTemp.instrPc = ctx.pc - 2
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
        resTemp.instrPc = ctx.pc - 2
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
        resTemp.instrPc = ctx.pc - 2
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
        resTemp.instrPc = ctx.pc - 2
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
        resTemp.instrPc = ctx.pc - 2
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
        resTemp.instrPc = ctx.pc - 2
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
        resTemp.instrPc = ctx.pc - 2
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
        resTemp.instrPc = ctx.pc - 2
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
        resTemp.instrPc = ctx.pc - 2
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
        resTemp.instrPc = ctx.pc - 2
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
        resTemp.instrPc = ctx.pc - 2
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
        resTemp.instrPc = ctx.pc - 2
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
        resTemp.instrPc = ctx.pc - 2
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
