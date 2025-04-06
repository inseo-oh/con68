// This file was automatically generated.
// Generated at 2025-04-05 21:25:18
package main

type instrSwap struct {
	// Fields
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

// ==========================================================================
// Decoder function
// ==========================================================================
func (ctx *clientContext) instrDecode() (res instr, err error) {
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
		} else {
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
