// Copyright (c) 2025, Oh Inseo (YJK) - Licensed under BSD-2-Clause
package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"slices"
)

//go:generate go run ./tool_autogen/ instr_autogen.go

func main() {
	addr := "127.0.0.1:6800"
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen to connection -- %v", err)
	}
	log.Printf("Started server at %s", addr)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept to connection -- %v", err)
			continue
		}
		log.Printf("New client connection from %s", conn.RemoteAddr().String())
		clientCtx := clientContext{
			conn:   conn,
			reader: bufio.NewReader(conn),
		}
		clientCtx.main()
		conn.Close()
	}

}

//==============================================================================
// State
//==============================================================================

type decodingContext = struct {
	ir uint16

	currInstrName string
	eaFields      [2]*ea
	eaRegs        [2]uint8
	opsize        opsize
}

type clientContext struct {
	decodingCtx    decodingContext
	lastExecutedIr uint16

	// Networking --------------------------------------------------------------
	conn   net.Conn
	reader *bufio.Reader
	closed bool

	// Registers ---------------------------------------------------------------
	dataRegs [8]uint32
	addrRegs [7]uint32
	a7ssp    uint32
	a7usp    uint32
	pc       uint32

	// SR/CCR flags ------------------------------------------------------------
	srT bool
	srS bool
	srI uint8

	ccrX bool
	ccrN bool
	ccrZ bool
	ccrV bool
	ccrC bool

	// Other flags -------------------------------------------------------------
	// "Trace" flags below control whether the trace event is sent to the client or not.
	// These are not related to 68000's tracing feature.
	traceExec      bool
	traceExc       bool
	stopped        bool
	inGroup0Or1Exc bool
}

//==============================================================================
// Operation size
//==============================================================================

type opsize uint8

const (
	opsizeNone = opsize(iota)
	opsizeByte
	opsizeWord
	opsizeLong
)

//==============================================================================
// SR/CCR
//==============================================================================

const (
	ccrFlagC = uint8(1 << 0)
	ccrFlagV = uint8(1 << 1)
	ccrFlagZ = uint8(1 << 2)
	ccrFlagN = uint8(1 << 3)
	ccrFlagX = uint8(1 << 4)
)

const (
	srFlagC = uint16(ccrFlagC)
	srFlagV = uint16(ccrFlagV)
	srFlagZ = uint16(ccrFlagZ)
	srFlagN = uint16(ccrFlagN)
	srFlagX = uint16(ccrFlagX)

	srFlagIOffset = uint16(8)
	srFlagIMask   = uint16(0x7 << srFlagIOffset)
	srFlagS       = uint16(1 << 13)
	srFlagT       = uint16(1 << 15)
)

func (ctx *clientContext) writeCcr(v uint8) {
	ctx.ccrX = (v & ccrFlagX) != 0
	ctx.ccrN = (v & ccrFlagN) != 0
	ctx.ccrZ = (v & ccrFlagZ) != 0
	ctx.ccrV = (v & ccrFlagV) != 0
	ctx.ccrC = (v & ccrFlagC) != 0
}

func (ctx *clientContext) readCcr() uint8 {
	result := uint8(0)
	if ctx.ccrX {
		result |= ccrFlagX
	}
	if ctx.ccrN {
		result |= ccrFlagN
	}
	if ctx.ccrZ {
		result |= ccrFlagZ
	}
	if ctx.ccrV {
		result |= ccrFlagV
	}
	if ctx.ccrC {
		result |= ccrFlagC
	}
	return result
}

func (ctx *clientContext) writeSr(v uint16) {
	ctx.writeCcr(uint8(v))
	ctx.srT = (v & srFlagT) != 0
	ctx.srS = (v & srFlagS) != 0
	ctx.srI = uint8((v & srFlagIMask) >> srFlagIOffset)
}

func (ctx *clientContext) readSr() uint16 {
	result := uint16(ctx.readCcr())
	if ctx.srT {
		result |= srFlagT
	}
	if ctx.srS {
		result |= srFlagS
	}
	result |= (uint16(ctx.srI) << srFlagIOffset) & srFlagIMask
	return result
}

func (ctx *clientContext) setNZFlagsB(v uint8) {
	ctx.ccrN = (v & 0x80) != 0
	ctx.ccrZ = v == 0
}
func (ctx *clientContext) setNZFlagsW(v uint16) {
	ctx.ccrN = (v & 0x8000) != 0
	ctx.ccrZ = v == 0
}
func (ctx *clientContext) setNZFlagsL(v uint32) {
	ctx.ccrN = (v & 0x80000000) != 0
	ctx.ccrZ = v == 0
}
func (ctx *clientContext) setNZFlags(v uint32, size opsize) {
	switch size {
	case opsizeByte:
		ctx.setNZFlagsB(uint8(v))
	case opsizeWord:
		ctx.setNZFlagsW(uint16(v))
	case opsizeLong:
		ctx.setNZFlagsL(v)
	default:
		panic("bad opsize")
	}
}
func (ctx *clientContext) clearVCFlags() {
	ctx.ccrV = false
	ctx.ccrC = false
}

// ==============================================================================
// Data and address registers
// ==============================================================================
func (ctx *clientContext) writeDregL(reg uint8, v uint32) {
	ctx.dataRegs[reg] = v
}
func (ctx *clientContext) writeDregW(reg uint8, v uint16) {
	ctx.dataRegs[reg] = ctx.dataRegs[reg] & ^uint32(0xffff) | (uint32(v) & uint32(0xffff))
}
func (ctx *clientContext) writeDregB(reg uint8, v uint8) {
	ctx.dataRegs[reg] = ctx.dataRegs[reg] & ^uint32(0xff) | (uint32(v) & uint32(0xff))
}
func (ctx *clientContext) writeDreg(reg uint8, size opsize, v uint32) {
	switch size {
	case opsizeByte:
		ctx.writeDregB(reg, uint8(v))
	case opsizeWord:
		ctx.writeDregW(reg, uint16(v))
	case opsizeLong:
		ctx.writeDregL(reg, v)
	default:
		panic("bad opsize")
	}
}
func (ctx *clientContext) readDregL(reg uint8) uint32 {
	return ctx.dataRegs[reg]
}
func (ctx *clientContext) readDregW(reg uint8) uint16 {
	return uint16(ctx.dataRegs[reg])
}
func (ctx *clientContext) readDregB(reg uint8) uint8 {
	return uint8(ctx.dataRegs[reg])
}
func (ctx *clientContext) readDreg(reg uint8, size opsize) uint32 {
	switch size {
	case opsizeByte:
		return uint32(ctx.readDregB(reg))
	case opsizeWord:
		return uint32(ctx.readDregW(reg))
	case opsizeLong:
		return ctx.readDregL(reg)
	default:
		panic("bad opsize")
	}
}

func (ctx *clientContext) getAreg(reg uint8) *uint32 {
	if reg != 7 {
		return &ctx.addrRegs[reg]
	} else if ctx.srS {
		return &ctx.a7ssp
	} else {
		return &ctx.a7usp
	}
}
func (ctx *clientContext) writeAregL(reg uint8, v uint32) {
	*ctx.getAreg(reg) = v
}
func (ctx *clientContext) writeAregW(reg uint8, v uint16) {
	*ctx.getAreg(reg) = signExtendWToL(v)
}
func (ctx *clientContext) writeAreg(reg uint8, size opsize, v uint32) {
	switch size {
	case opsizeWord:
		ctx.writeAregW(reg, uint16(v))
	case opsizeLong:
		ctx.writeAregL(reg, v)
	default:
		panic("bad opsize")
	}
}
func (ctx *clientContext) readAreg(reg uint8) uint32 {
	return *ctx.getAreg(reg)
}

func (ctx *clientContext) decrementAreg(reg uint8, size opsize) uint32 {
	incr := int32(0)

	switch size {
	case opsizeByte:
		if reg == 7 {
			// SP is decremented by 2 even if it's byte sized, to keep it aligned.
			incr = -2
		} else {
			incr = -1
		}
	case opsizeWord:
		incr = -2
	case opsizeLong:
		incr = -4
	default:
		panic("bad opsize")
	}
	val := ctx.readAreg(reg)
	val += uint32(incr)
	ctx.writeAregL(reg, val)
	return val
}

func (ctx *clientContext) incrementAreg(reg uint8, size opsize) uint32 {
	incr := uint32(0)

	switch size {
	case opsizeByte:
		if reg == 7 {
			// SP is incremented by 2 even if it's byte sized, to keep it aligned.
			incr = 2
		} else {
			incr = 1
		}
	case opsizeWord:
		incr = 2
	case opsizeLong:
		incr = 4
	default:
		panic("bad opsize")
	}
	val := ctx.readAreg(reg)
	val += uint32(incr)
	ctx.writeAregL(reg, val)
	return val
}

//==============================================================================
// Effective addressing
//==============================================================================

type eamode uint8

const (
	eamodeDreg           = eamode(iota) // Data register direct [Dn]
	eamodeAreg                          // Address register direct [An]
	eamodeAregInd                       // Address register indirect [(An)]
	eamodeAregIndPostinc                // Address register indirect + postincrement [(An)+]
	eamodeAregIndPredec                 // Address register indirect + predecrement [-(An)]
	eamodeAregIndDisp                   // Address register indirect + displacement [(d8, An)]
	eamodeAregIndIndex                  // Address register indirect + index [(d8, An, Xn)]
	eamodeAbsW                          // Absolute (16-bit sign-extended) [xxx.w]
	eamodeAbsL                          // Absolute (32-bit) [xxx.l]
	eamodePcIndDisp                     // PC indirect + displacement [(d8, PC)]
	eamodePcIndIndex                    // PC indirect + index [(d8, PC)]
	eamodeImm                           // Immediate (#xxx)
)

type ea struct {
	mode eamode

	// Meaning of this field depends on the mode:
	// - Register direct/indirect modes: Index of the register
	// - PC indirect mode: The PC address
	// - Absolute modes: Absolute address. It is final 32-bit address, regardless of the mode
	// - Immediate mode: The immediate value
	//
	// To avoid confusion, do not read from this field directly.
	// Use accessor functions instead, which will also check the mode to prevent bugs.
	_val uint32

	// Displacement is only valid for register indirect with displacement or index modes:
	// - Register indirect w/ displacement: 16-bit index
	// - Register indirect w/ index: 8-bit index
	//
	// In both cases displacement is stored as sign-extended 32-bit value.
	// Again, do not read from this directly.
	_disp uint32

	// For indexing modes ------------------------------------------------------
	indexRegType regType
	indexSize    opsize
	indexReg     uint8
}

func (ea ea) reg() uint8 {
	switch ea.mode {
	case eamodeDreg, eamodeAreg, eamodeAregInd, eamodeAregIndPostinc, eamodeAregIndPredec, eamodeAregIndDisp, eamodeAregIndIndex:
		return uint8(ea._val)
	}
	panic("called with non-applicable EA mode")
}
func (ea ea) absAddr() uint32 {
	switch ea.mode {
	case eamodeAbsW, eamodeAbsL:
		return ea._val
	}
	panic("called with non-applicable EA mode")
}
func (ea ea) pcAddress() uint32 {
	switch ea.mode {
	case eamodePcIndDisp, eamodePcIndIndex:
		return ea._val
	}
	panic("called with non-applicable EA mode")
}
func (ea ea) imm() uint32 {
	switch ea.mode {
	case eamodeImm:
		return ea._val
	}
	panic("called with non-applicable EA mode")
}
func (ea ea) disp() uint32 {
	switch ea.mode {
	case eamodeAregIndDisp, eamodeAregIndIndex, eamodePcIndDisp, eamodePcIndIndex:
		return ea._disp
	}
	panic("called with non-applicable EA mode")
}
func (ea ea) ToString() string {
	switch ea.mode {
	case eamodeDreg:
		return fmt.Sprintf("d%d", ea.reg())
	case eamodeAreg:
		return fmt.Sprintf("a%d", ea.reg())
	case eamodeAregInd:
		return fmt.Sprintf("(a%d)", ea.reg())
	case eamodeAregIndPostinc:
		return fmt.Sprintf("(a%d)+", ea.reg())
	case eamodeAregIndPredec:
		return fmt.Sprintf("-(a%d)", ea.reg())
	case eamodeAregIndDisp:
		return fmt.Sprintf("(%d, a%d)", ea.disp(), ea.reg())
	case eamodeAregIndIndex:
		return fmt.Sprintf("(%d, a%d, %s%d)", ea.disp(), ea.reg(), ea.indexRegType.ToString(), ea.indexReg)
	case eamodePcIndDisp:
		return fmt.Sprintf("(%d, pc)", ea.disp())
	case eamodePcIndIndex:
		return fmt.Sprintf("(%d, pc, %s%d)", ea.disp(), ea.indexRegType.ToString(), ea.indexReg)
	case eamodeAbsW, eamodeAbsL:
		return fmt.Sprintf("$%08X", ea.absAddr())
	case eamodeImm:
		return fmt.Sprintf("#$%08X", ea.imm())
	}
	panic("bad eamode")
}

func (ctx *clientContext) memAddrOfIndexedEa(ea ea) uint32 {
	baseAddr := uint32(0)
	switch ea.mode {
	case eamodeAregIndIndex:
		baseAddr = ctx.readAreg(ea.reg())
	case eamodePcIndIndex:
		baseAddr = ea.pcAddress()
	default:
		panic("called with non-applicable EA mode")
	}
	offset := uint32(0)
	switch ea.indexRegType {
	case regTypeAddr:
		offset = ctx.readAreg(ea.indexReg)
	case regTypeData:
		offset = ctx.readDregL(ea.indexReg)
	default:
		panic("bad regType value")
	}
	switch ea.indexSize {
	case opsizeLong:
		break
	case opsizeWord:
		offset = signExtendWToL(uint16(offset))
	}
	offset += ea.disp()
	return baseAddr + offset
}

func (ctx *clientContext) memAddrOfEa(ea ea, size opsize) uint32 {
	switch ea.mode {
	case eamodeAregInd:
		reg := ea.reg()
		return ctx.readAreg(reg)
	case eamodeAregIndPostinc:
		reg := ea.reg()
		addr := ctx.readAreg(reg)
		ctx.incrementAreg(reg, size)
		return addr
	case eamodeAregIndPredec:
		reg := ea.reg()
		return ctx.decrementAreg(reg, size)
	case eamodeAregIndDisp:
		reg := ea.reg()
		disp := ea.disp()
		return ctx.readAreg(reg) + disp
	case eamodePcIndDisp:
		disp := ea.disp()
		return ea.pcAddress() + disp
	case eamodeAregIndIndex, eamodePcIndIndex:
		return ctx.memAddrOfIndexedEa(ea)
	case eamodeAbsW, eamodeAbsL:
		return ea.absAddr()
	case eamodeDreg, eamodeAreg, eamodeImm:
		panic("attempted to get memory address of what isn't memory address operand")
	}
	fmt.Printf("%d\n", ea.mode)
	panic("bad eamode")
}

func (ctx *clientContext) readEa(ea ea, size opsize) (uint32, error) {
	switch ea.mode {
	case eamodeDreg:
		return ctx.readDreg(ea.reg(), size), nil
	case eamodeAreg:
		if size != opsizeLong {
			panic("An mode can only be read with long size")
		}
		return ctx.readAreg(ea.reg()), nil
	case eamodeImm:
		return ea.imm(), nil
	case eamodeAregInd, eamodeAregIndPostinc, eamodeAregIndPredec, eamodeAregIndDisp, eamodeAregIndIndex, eamodeAbsW, eamodeAbsL, eamodePcIndDisp, eamodePcIndIndex:
		addr := ctx.memAddrOfEa(ea, size)
		fc := ctx.getFuncCode(false)
		return ctx.readMem(addr, fc, size)
	}
	panic("bad eamode")
}

func (ctx *clientContext) writeEa(ea ea, size opsize, v uint32) error {
	switch ea.mode {
	case eamodeDreg:
		ctx.writeDreg(ea.reg(), size, v)
		return nil
	case eamodeAreg:
		ctx.writeAreg(ea.reg(), size, v)
		return nil
	case eamodeImm:
		panic("attempted to write to immediate")
	case eamodeAregInd, eamodeAregIndPostinc, eamodeAregIndPredec, eamodeAregIndDisp, eamodeAregIndIndex, eamodeAbsW, eamodeAbsL:
		addr := ctx.memAddrOfEa(ea, size)
		fc := ctx.getFuncCode(false)
		return ctx.writeMem(addr, fc, size, v)
	}
	panic("bad eamode")
}

func (ctx *clientContext) readModifyWriteEa(ea ea, size opsize, modify func(uint32) uint32) error {
	switch ea.mode {
	case eamodeDreg:
		v := ctx.readDreg(ea.reg(), size)
		v = modify(v)
		ctx.writeDreg(ea.reg(), size, v)
		return nil
	case eamodeAreg:
		if size != opsizeLong {
			panic("An mode can only be read with long size")
		}
		v := ctx.readAreg(ea.reg())
		v = modify(v)
		ctx.writeAreg(ea.reg(), size, v)
		return nil
	case eamodeImm:
		panic("attempted to write to immediate")
	case eamodeAregInd, eamodeAregIndPostinc, eamodeAregIndPredec, eamodeAregIndDisp, eamodeAregIndIndex, eamodeAbsW, eamodeAbsL:
		addr := ctx.memAddrOfEa(ea, size)
		fc := ctx.getFuncCode(false)
		v, err := ctx.readMem(addr, fc, size)
		if err != nil {
			return err
		}
		return ctx.writeMem(addr, fc, size, v)
	}
	panic("bad eamode")
}

//==============================================================================
// Conditions
//==============================================================================

// WARNING: These values also correspond to condition field in instructions!
// DO NOT CHANGE WITHOUT A REASON.
type cond uint8

const (
	condT  = cond(0x0) // True (**Not applicable to Bcc!**)
	condF  = cond(0x1) // False (**Not applicable to Bcc!**)
	condHi = cond(0x2) // High
	condLs = cond(0x3) // Low or Same
	condCc = cond(0x4) // Carry Clear
	condCs = cond(0x5) // Carry Set
	condNe = cond(0x6) // Not Equal
	condEq = cond(0x7) // Equal
	condVc = cond(0x8) // Overflow Clear
	condVs = cond(0x9) // Overflow Set
	condPl = cond(0xa) // Plus
	condMi = cond(0xb) // Minus
	condGe = cond(0xc) // Greater or Equal
	condLt = cond(0xd) // Less Than
	condGt = cond(0xe) // Greater Than
	condLe = cond(0xf) // Less or Equal
)

func (cond cond) ToString() string {
	condStrs := [16]string{
		"t",
		"f",
		"hi",
		"ls",
		"cc",
		"cs",
		"ne",
		"eq",
		"vc",
		"vs",
		"pl",
		"mi",
		"ge",
		"lt",
		"gt",
		"le",
	}
	return condStrs[cond]
}

func (ctx *clientContext) testCond(cond cond) bool {
	ccrN := ctx.ccrN
	ccrZ := ctx.ccrZ
	ccrV := ctx.ccrV
	ccrC := ctx.ccrC
	switch cond {
	case condT:
		return true
	case condF:
		return false
	case condHi:
		return !ccrC && !ccrZ
	case condLs:
		return ccrC || ccrZ
	case condCc:
		return !ccrC
	case condCs:
		return ccrC
	case condNe:
		return !ccrZ
	case condEq:
		return ccrZ
	case condVc:
		return !ccrV
	case condVs:
		return ccrV
	case condPl:
		return !ccrN
	case condMi:
		return ccrN
	case condGe:
		return (ccrN && ccrV) || (!ccrN && !ccrV)
	case condLt:
		return (ccrN && !ccrV) || (!ccrN && ccrV)
	case condGt:
		return (ccrN && ccrV && !ccrZ) || (!ccrN && !ccrV && !ccrZ)
	case condLe:
		return ccrZ || (ccrN && !ccrV) || (!ccrN && ccrV)
	}
	panic("bad cond value")
}

//==============================================================================
// Instruction interface
//==============================================================================

type instr interface {
	disasm() string
	exec(ctx *clientContext) error
}

//==============================================================================
// Instruction decoding
//==============================================================================

func (ctx *clientContext) fetchInstrW() (uint16, error) {
	fc := ctx.getFuncCode(true)
	res, err := ctx.readMemW(ctx.pc, fc)
	if err != nil {
		return 0, err
	}
	ctx.pc += 2
	return res, nil
}
func (ctx *clientContext) fetchInstrL() (uint32, error) {
	fc := ctx.getFuncCode(true)
	res, err := ctx.readMemL(ctx.pc, fc)
	if err != nil {
		return 0, err
	}
	ctx.pc += 4
	return res, nil
}

// These are helper functions that extract the raw field value.
//
// I could've made those take directly from the IR, but then golang formatter decides it's a bit too long and breaks into multiple lines.
// Besides, this is easier to read anyway.
func fieldSizeType1(x uint16) uint8 { return uint8(((x) & (0x3 << 6)) >> 6) }   // ........XX......
func fieldSizeType2(x uint16) uint8 { return uint8(((x) & (0x3 << 12)) >> 12) } // ..XX............
func fieldSizeType3(x uint16) uint8 { return uint8(((x) & (0x1 << 6)) >> 6) }   // .........X......
func fieldCond(x uint16) uint8      { return uint8(((x) & (0xf << 8)) >> 8) }   // ....XXXX........
func fieldVector(x uint16) uint8    { return uint8(((x) & (0xf << 0)) >> 0) }   // ............XXXX
func fieldImm8(x uint16) uint8      { return uint8(((x) & (0xff << 0)) >> 0) }  // ........XXXXXXXX
func fieldImm3(x uint16) uint8      { return uint8(((x) & (0x7 << 9)) >> 9) }   // ....XXX.........
func fieldRegX(x uint16) uint8      { return uint8(((x) & (0x7 << 9)) >> 9) }   // ....XXX.........
func fieldRegY(x uint16) uint8      { return uint8(((x) & (0x7 << 0)) >> 0) }   // .............XXX
func fieldModeX(x uint16) uint8     { return uint8(((x) & (0x7 << 6)) >> 6) }   // .......XXX......
func fieldModeY(x uint16) uint8     { return uint8(((x) & (0x7 << 3)) >> 3) }   // ..........XXX...

// NOTE: These instruction decoding functions get referenced by the auto-generated instruction decoder code.

// So far I found THREE different ways to encode size. Thanks Motorola :D
//            | 00 | 01 | 10 | 11 | Note
// SIZE_TYPE1 | B  | W  | L  |    |                 |
// SIZE_TYPE2 |    | B  | L  | W  |                 |
// SIZE_TYPE3 | W  | L  |    |    | This uses 1-bit |

func (ctx *clientContext) decodeFieldSizeType1() (opsize, bool) {
	switch fieldSizeType1(ctx.decodingCtx.ir) {
	case 0x0:
		return opsizeByte, true
	case 0x1:
		return opsizeWord, true
	case 0x2:
		return opsizeLong, true
	default:
		return 0, false
	}
}
func (ctx *clientContext) decodeFieldSizeType2() (opsize, bool) {
	switch fieldSizeType2(ctx.decodingCtx.ir) {
	case 0x1:
		return opsizeByte, true
	case 0x3:
		return opsizeWord, true
	case 0x2:
		return opsizeLong, true
	default:
		return 0, false
	}
}
func (ctx *clientContext) decodeFieldSizeType3() (opsize, bool) {
	switch fieldSizeType3(ctx.decodingCtx.ir) {
	case 0x0:
		return opsizeWord, true
	case 0x1:
		return opsizeLong, true
	default:
		return 0, false
	}
}
func (ctx *clientContext) decodeFieldCond() (cond, bool) {
	res := cond(fieldCond(ctx.decodingCtx.ir))
	if (ctx.decodingCtx.currInstrName == "bcc") && ((res == condT) || (res == condF)) {
		//  Bcc doesn't allow T or F condition
		return 0, false
	}
	return res, true
}
func (ctx *clientContext) decodeFieldImm3() (uint8, bool) {
	return fieldImm3(ctx.decodingCtx.ir), true
}
func (ctx *clientContext) decodeFieldImm8() (uint8, bool) {
	return fieldImm8(ctx.decodingCtx.ir), true
}
func (ctx *clientContext) decodeFieldVector() (uint8, bool) {
	return fieldVector(ctx.decodingCtx.ir), true
}
func (ctx *clientContext) decodeFieldRegX() (uint8, bool) {
	return fieldRegX(ctx.decodingCtx.ir), true
}
func (ctx *clientContext) decodeFieldRegY() (uint8, bool) {
	return fieldRegY(ctx.decodingCtx.ir), true
}
func (ctx *clientContext) decodeEaField(mode, reg uint8) (eamode, bool) {
	switch mode {
	case 0:
		return eamodeDreg, true
	case 1:
		return eamodeAreg, true
	case 2:
		return eamodeAregInd, true
	case 3:
		return eamodeAregIndPostinc, true
	case 4:
		return eamodeAregIndPredec, true
	case 5:
		return eamodeAregIndDisp, true
	case 6:
		return eamodeAregIndIndex, true
	case 7:
		switch reg {
		case 0:
			return eamodeAbsW, true
		case 1:
			return eamodeAbsL, true
		case 2:
			return eamodePcIndDisp, true
		case 3:
			return eamodePcIndIndex, true
		case 4:
			return eamodeImm, true
		}
	}
	return 0, false
}

// EA decoders return pointer to new EA, because they are only partially initialized here.
// After all other fields and extension words are decoded, then we can decode the EA.
func (ctx *clientContext) decodeFieldEa1() (*ea, bool) {
	mode := fieldModeY(ctx.decodingCtx.ir)
	reg := fieldRegY(ctx.decodingCtx.ir)
	eamode, ok := ctx.decodeEaField(mode, reg)
	if !ok {
		return nil, false
	}
	ea := new(ea)
	ea.mode = eamode
	ctx.decodingCtx.eaFields[0] = ea
	ctx.decodingCtx.eaRegs[0] = reg
	return ea, true
}
func (ctx *clientContext) decodeFieldEa2() (*ea, bool) {
	mode := fieldModeX(ctx.decodingCtx.ir)
	reg := fieldRegX(ctx.decodingCtx.ir)
	eamode, ok := ctx.decodeEaField(mode, reg)
	if !ok {
		return nil, false
	}
	ea := new(ea)
	ea.mode = eamode
	ctx.decodingCtx.eaFields[1] = ea
	ctx.decodingCtx.eaRegs[1] = reg
	return ea, true
}
func (ctx *clientContext) checkEaModes(ea1Modes []eamode, ea2Modes []eamode) bool {
	if len(ea1Modes) != 0 {
		field := ctx.decodingCtx.eaFields[0]
		if field == nil {
			panic("ea1 field must be present")
		}
		ok := slices.Contains(ea1Modes, field.mode)
		if !ok {
			return false
		}
	}
	if len(ea2Modes) != 0 {
		field := ctx.decodingCtx.eaFields[1]
		if field == nil {
			panic("ea2 field must be present")
		}
		ok := slices.Contains(ea2Modes, field.mode)
		if !ok {
			return false
		}
	}
	return true
}
func (ctx *clientContext) decodeEa() error {
	for i := range 2 {
		isDisp := false
		isIndex := false
		dest := ctx.decodingCtx.eaFields[i]
		regField := ctx.decodingCtx.eaRegs[i]
		if dest == nil {
			continue
		}
		switch dest.mode {
		case eamodeDreg, eamodeAreg, eamodeAregInd, eamodeAregIndPostinc, eamodeAregIndPredec:
			dest._val = uint32(regField)
		case eamodeAregIndDisp:
			dest._val = uint32(regField)
			isDisp = true
		case eamodeAregIndIndex:
			dest._val = uint32(regField)
			isIndex = true
		case eamodeAbsW:
			v, err := ctx.fetchInstrW()
			if err != nil {
				return err
			}
			dest._val = signExtendWToL(v)
		case eamodeAbsL:
			v, err := ctx.fetchInstrL()
			if err != nil {
				return err
			}
			dest._val = v
		case eamodePcIndDisp:
			dest._val = ctx.pc
			isDisp = true
		case eamodePcIndIndex:
			dest._val = ctx.pc
			isIndex = true
		case eamodeImm:
			switch ctx.decodingCtx.opsize {
			case opsizeByte:
				v, err := ctx.fetchInstrW()
				if err != nil {
					return err
				}
				dest._val = uint32(v & 0xff)
			case opsizeWord:
				v, err := ctx.fetchInstrW()
				if err != nil {
					return err
				}
				dest._val = uint32(v)
			case opsizeLong:
				v, err := ctx.fetchInstrL()
				if err != nil {
					return err
				}
				dest._val = v
			default:
				panic("bad opsize")
			}

		default:
			panic("bad opmode")
		}

		if isDisp {
			v, err := ctx.fetchInstrW()
			if err != nil {
				return err
			}
			dest._disp = signExtendWToL(v)
		}
		if isIndex {
			v, err := ctx.fetchInstrW()
			if err != nil {
				return err
			}
			if v&(1<<15) != 0 {
				dest.indexRegType = regTypeAddr
			} else {
				dest.indexRegType = regTypeData
			}
			if v&(1<<11) != 0 {
				dest.indexSize = opsizeLong
			} else {
				dest.indexSize = opsizeWord
			}
			dest.indexReg = uint8((v >> 12) & 0x7)
			dest._disp = signExtendBToL(uint8(v))
		}
	}
	return nil
}

// Extension word decoding

func (ctx *clientContext) decodeXwordBranchOff() (uint32, error) {
	if v := fieldImm8(ctx.decodingCtx.ir); v != 0 {
		return signExtendBToL(v), nil
	}
	v, err := ctx.fetchInstrW()
	if err != nil {
		return 0, err
	}
	return signExtendWToL(v), nil
}
func (ctx *clientContext) decodeXwordImm8() (uint8, error) {
	v, err := ctx.fetchInstrW()
	if err != nil {
		return 0, err
	}
	return uint8(v), nil
}
func (ctx *clientContext) decodeXwordImm16() (uint16, error) {
	return ctx.fetchInstrW()
}
func (ctx *clientContext) decodeXwordImm32() (uint32, error) {
	return ctx.fetchInstrL()
}
func (ctx *clientContext) decodeXwordImm() (uint32, error) {
	switch ctx.decodingCtx.opsize {
	case opsizeByte:
		v, err := ctx.fetchInstrW()
		if err != nil {
			return 0, err
		}
		return uint32(v & 0xff), nil
	case opsizeWord:
		v, err := ctx.fetchInstrW()
		if err != nil {
			return 0, err
		}
		return uint32(v), nil
	case opsizeLong:
		v, err := ctx.fetchInstrL()
		if err != nil {
			return 0, err
		}
		return v, nil
	default:
		panic("bad opsize")
	}
}

//==============================================================================
// Exception and vector
//==============================================================================

type exc uint8

const (
	excResetSsp                  = exc(0x0)  // Vector for SSP value after reset. Not an actual exception.
	excResetPc                   = exc(0x1)  // Vector for PC value after reset. Not an actual exception.
	excBusError                  = exc(0x2)  // Bus error (System bus asserted BERR signal)
	excAddressError              = exc(0x3)  // Address error (Non-word aligned address)
	excIllegalInstr              = exc(0x4)  // Illegal instruction
	excZeroDivide                = exc(0x5)  // Divide by zero
	excChk                       = exc(0x6)  // Exception caused by CHK instruction
	excTrapv                     = exc(0x7)  // Exception caused by TRAPV instruction
	excPrivilegeViolation        = exc(0x8)  // Privilege violation
	excTrace                     = exc(0x9)  // Tracing
	excLineA                     = exc(0xa)  // Line A emulator
	excLineF                     = exc(0xb)  // Line F emulator
	excSpuriousInterrupt         = exc(0x18) // Spurious interrupt
	excLevel1InterruptAutovector = exc(0x19) // Level 1 interrupt autovector
	excLevel2InterruptAutovector = exc(0x1a) // Level 2 interrupt autovector
	excLevel3InterruptAutovector = exc(0x1b) // Level 3 interrupt autovector
	excLevel4InterruptAutovector = exc(0x1c) // Level 4 interrupt autovector
	excLevel5InterruptAutovector = exc(0x1d) // Level 5 interrupt autovector
	excLevel6InterruptAutovector = exc(0x1e) // Level 6 interrupt autovector
	excLevel7InterruptAutovector = exc(0x1f) // Level 7 interrupt autovector
	excTrapVectorStart           = exc(0x20) // Start offset of TRAP vectors
)

type excError struct {
	memExcAddr  uint32
	ir          uint16
	memExcFlags uint8
	exc         exc
}

func (e excError) Error() string {
	return fmt.Sprintf("68000 Exception %#x", e.exc)
}

func (e excError) isGroup0Or1Exc() bool {
	switch e.exc {
	case excZeroDivide:
	case excChk:
	case excTrapv:
	case excTrace:
		return true
	default:
		return excTrapVectorStart <= e.exc
	}
	return false
}

func (ctx *clientContext) memExcError(exc exc, addr uint32, fc fc, dir busDir) excError {
	flags := uint8(fc)
	// I/N
	if ctx.inGroup0Or1Exc {
		flags |= 1 << 3
	}
	// R/W
	if dir == busDirRead {
		flags |= 1 << 4
	}
	return excError{
		exc:         exc,
		ir:          ctx.decodingCtx.ir,
		memExcFlags: flags,
		memExcAddr:  addr,
	}
}

func (ctx *clientContext) fetchVector(exc exc) (uint32, error) {
	isProgram := false
	switch exc {
	case excResetPc:
	case excResetSsp:
		isProgram = true
	}
	fc := ctx.getFuncCode(isProgram)
	return ctx.readMemL(uint32(exc)*4, fc)
}

func (ctx *clientContext) beginExc(err excError) error {
	// NOTE: If another exception occurs(which would be either address or bus error), we MUST handle it here.
	// And to avoid stack overflow, we shouldn't even call this function recursively, because it is possible to cause infinite bus error loop.
	currentErr := err

	for {
		ctx.inGroup0Or1Exc = err.isGroup0Or1Exc()
		newPc, err := ctx.handleExc(currentErr)
		if err != nil {
			if excErr, isExcErr := err.(excError); isExcErr {
				if (excErr.exc == excAddressError) && (currentErr.exc == excAddressError) {
					// Double address error
					if err := ctx.eventTraceExcMem(excErr.exc, ctx.pc, excErr.ir, excErr.memExcAddr, excErr.memExcFlags); err != nil {
						return err
					}
					return err
				} else {
					// Begin new exception
					currentErr = excErr
					continue
				}
			} else {
				// Not an exception error (e.g. network error)
				return err
			}
		}
		ctx.inGroup0Or1Exc = false
		ctx.pc = newPc
		break
	}
	return nil
}

// Internal helper
func (ctx *clientContext) handleExc(err excError) (uint32, error) {
	pc := ctx.pc
	isMemErr := (err.exc == excBusError) || (err.exc == excAddressError)
	if ctx.traceExc {
		if isMemErr {
			if err := ctx.eventTraceExcMem(err.exc, pc, err.ir, err.memExcAddr, err.memExcFlags); err != nil {
				return 0, err
			}
		} else {
			if err := ctx.eventTraceExc(err.exc, pc); err != nil {
				return 0, err
			}
		}
	}
	oldSr := ctx.readSr()
	ctx.srS = true
	ctx.srT = false
	newPc := uint32(0)
	// TODO: Should we set srI as well?
	if v, err := ctx.fetchVector(err.exc); err != nil {
		return 0, err
	} else {
		newPc = v
	}
	if err := ctx.pushL(pc); err != nil {
		return 0, err
	}
	if err := ctx.pushW(oldSr); err != nil {
		return 0, err
	}
	if isMemErr {
		w := err.ir
		if err := ctx.pushW(w); err != nil {
			return 0, err
		}
		if err := ctx.pushL(err.memExcAddr); err != nil {
			return 0, err
		}
		// Based on JSON tests I was using, seems like unused bits come from the IR?
		w &= ^uint16(0x1f)
		w |= uint16(err.memExcFlags) & 0x1f
		if err := ctx.pushW(w); err != nil {
			return 0, err
		}
	}
	return newPc, nil
}

//==============================================================================
// Stack
//==============================================================================

func (ctx *clientContext) pushW(v uint16) error {
	fc := ctx.getFuncCode(false)
	addr := ctx.decrementAreg(7, opsizeWord)
	return ctx.writeMemW(addr, fc, v)
}

func (ctx *clientContext) pushL(v uint32) error {
	fc := ctx.getFuncCode(false)
	addr := ctx.decrementAreg(7, opsizeLong)
	return ctx.writeMemL(addr, fc, v)
}

func (ctx *clientContext) popW() (uint16, error) {
	fc := ctx.getFuncCode(false)
	addr := ctx.readAreg(7)
	if v, err := ctx.readMemW(addr, fc); err != nil {
		return 0, err
	} else {
		ctx.incrementAreg(7, opsizeWord)
		return v, nil
	}
}

func (ctx *clientContext) popL() (uint32, error) {
	fc := ctx.getFuncCode(false)
	addr := ctx.readAreg(7)
	if v, err := ctx.readMemL(addr, fc); err != nil {
		return 0, err
	} else {
		ctx.incrementAreg(7, opsizeLong)
		return v, nil
	}
}

//==============================================================================
// Memory bus
//==============================================================================

type busDir uint8

const (
	busDirRead = busDir(iota)
	busDirWrite
)

func (ds netDs) valueMask() uint16 {
	switch ds {
	case netDsUpper:
		return 0xff00
	case netDsLower:
		return 0x00ff
	case netDsBoth:
		return 0xffff
	default:
		log.Panicf("%d is not a valid DS value", ds)
		return 0
	}
}

type fc uint8

const (
	fcFlagData    = fc(1 << 0)
	fcFlagProgram = fc(1 << 1)
	fcFlagSuper   = fc(1 << 2)

	fcUserData     = fcFlagData
	fcUserProgram  = fcFlagProgram
	fcSuperData    = fcFlagSuper | fcFlagData
	fcSuperProgram = fcFlagSuper | fcFlagProgram
	fcCpuSpace     = fcFlagSuper | fcFlagProgram | fcFlagData
)

func (ctx *clientContext) getFuncCode(isProgram bool) fc {
	result := fc(0)
	// FC0~FC1
	if isProgram {
		result |= fcFlagProgram
	} else {
		result |= fcFlagData
	}
	// FC2
	if ctx.srS {
		result |= fcFlagSuper
	}
	return result
}

func (ctx *clientContext) readBus(addr uint32, ds netDs, fc fc) (uint16, error) {
	if (addr & 0x1) != 0 {
		return 0, ctx.memExcError(excAddressError, addr, fc, busDirRead)
	}
	addr &= ^uint32(0xff000000) // Limit to 24-bit
	if err := ctx.eventAddrAsserted(addr); err != nil {
		return 0, err
	}
	return ctx.eventReadBus(ds)
}
func (ctx *clientContext) writeBus(addr uint32, ds netDs, fc fc, v uint16) error {
	if (addr & 0x1) != 0 {
		return ctx.memExcError(excAddressError, addr, fc, busDirWrite)
	}
	addr &= ^uint32(0xff000000) // Limit to 24-bit
	if err := ctx.eventAddrAsserted(addr); err != nil {
		return err
	}
	return ctx.eventWriteBus(ds, v)
}
func (ctx *clientContext) readMemL(addr uint32, fc fc) (uint32, error) {
	result := uint32(0)
	if v, err := ctx.readBus(addr, netDsBoth, fc); err != nil {
		return 0, err
	} else {
		result = uint32(v) << 16
	}
	if v, err := ctx.readBus(addr+2, netDsBoth, fc); err != nil {
		return 0, err
	} else {
		result |= uint32(v)
	}
	return result, nil
}
func (ctx *clientContext) readMemW(addr uint32, fc fc) (uint16, error) {
	return ctx.readBus(addr, netDsBoth, fc)
}
func (ctx *clientContext) readMemB(addr uint32, fc fc) (uint8, error) {
	var ds netDs
	var shift uint32

	if (addr % 2) != 0 {
		ds = netDsLower
		shift = 0
	} else {
		ds = netDsUpper
		shift = 8
	}
	addr &= ^uint32(0x1)
	if v, err := ctx.readBus(addr, ds, fc); err != nil {
		return 0, err
	} else {
		return uint8(v >> shift), nil
	}
}
func (ctx *clientContext) readMem(addr uint32, fc fc, size opsize) (uint32, error) {
	switch size {
	case opsizeByte:
		if v, err := ctx.readMemB(addr, fc); err != nil {
			return 0, err
		} else {
			return uint32(v), nil
		}
	case opsizeWord:
		if v, err := ctx.readMemW(addr, fc); err != nil {
			return 0, err
		} else {
			return uint32(v), nil
		}
	case opsizeLong:
		return ctx.readMemL(addr, fc)
	default:
		panic("bad opsize")
	}
}
func (ctx *clientContext) writeMemL(addr uint32, fc fc, v uint32) error {
	if err := ctx.writeBus(addr, netDsBoth, fc, uint16(v>>16)); err != nil {
		return err
	}
	if err := ctx.writeBus(addr+2, netDsBoth, fc, uint16(v)); err != nil {
		return err
	}
	return nil
}
func (ctx *clientContext) writeMemW(addr uint32, fc fc, v uint16) error {
	return ctx.writeBus(addr, netDsBoth, fc, v)
}
func (ctx *clientContext) writeMemB(addr uint32, fc fc, v uint8) error {
	var ds netDs
	var shift uint32

	if (addr % 2) != 0 {
		ds = netDsLower
		shift = 0
	} else {
		ds = netDsUpper
		shift = 8
	}
	addr &= ^uint32(0x1)
	return ctx.writeBus(addr, ds, fc, uint16(v)<<shift)
}
func (ctx *clientContext) writeMem(addr uint32, fc fc, size opsize, v uint32) error {
	switch size {
	case opsizeByte:
		return ctx.writeMemB(addr, fc, uint8(v))
	case opsizeWord:
		return ctx.writeMemW(addr, fc, uint16(v))
	case opsizeLong:
		return ctx.writeMemL(addr, fc, v)
	default:
		panic("bad opsize")
	}
}

//==============================================================================
// Networking
//==============================================================================

// Every message(request or response) starts with header byte telling what kind of message it's sending
// Note that commands always come from the client
type netOpbyte uint8

const (
	// 0x - Response type.
	// Every response starts with this byte,
	netOpbyteAck  = netOpbyte(0x00) // Acknowledged
	netOpbyteFail = netOpbyte(0x01) // Failed

	// 1x - General commands
	netOpbyteBye          = netOpbyte(0x10) // Close the connection
	netOpbyteUnstop       = netOpbyte(0x11) // Unstop the CPU
	netOpbyteIsStopped    = netOpbyte(0x12) // Is the CPU stopped?
	netOpbyteTraceExecOn  = netOpbyte(0x13) // Trace Execution - Enable
	netOpbyteTraceExecOff = netOpbyte(0x14) // Trace Execution - Disable
	netOpbyteTraceExcOn   = netOpbyte(0x15) // Trace Exception - Enable
	netOpbyteTraceExcOff  = netOpbyte(0x16) // Trace Exception - Disable
	netOpbyteTick         = netOpbyte(0x1f) // Run the CPU for a tick

	// 2x - CPU state manipulation commands
	netOpbyteDregWrite = netOpbyte(0x20) // Data Register Write
	netOpbyteDregRead  = netOpbyte(0x21) // Data Register Read
	netOpbyteAregWrite = netOpbyte(0x22) // Address Register Write
	netOpbyteAregRead  = netOpbyte(0x23) // Address Register Read
	netOpbyteSspWrite  = netOpbyte(0x24) // A7(SSP) write
	netOpbyteSspRead   = netOpbyte(0x25) // A7(SSP) read
	netOpbyteUspWrite  = netOpbyte(0x26) // A7(USP) write
	netOpbyteUspRead   = netOpbyte(0x27) // A7(USP) read
	netOpbytePcWrite   = netOpbyte(0x28) // PC write
	netOpbytePcRead    = netOpbyte(0x29) // PC read
	netOpbyteSrWrite   = netOpbyte(0x2a) // SR write
	netOpbyteSrRead    = netOpbyte(0x2b) // SR Read

	// 8x - Server events
	// When client receives one of these, it should respond to it accordingly.
	netOpbyteEventAddrAsserted = netOpbyte(0x80) // Address asserted
	netOpbyteEventReadBus      = netOpbyte(0x81) // Read from last asserted address
	netOpbyteEventWriteBus     = netOpbyte(0x82) // Write to last asserted address
	netOpbyteEventReset        = netOpbyte(0x83) // RESET asserted
	netOpbyteEventTraceExec    = netOpbyte(0x84) // Event for Trace Execution
	netOpbyteEventTraceExc     = netOpbyte(0x85) // Event for Trace Exception (Non-memory exception)
	netOpbyteEventTraceExcMem  = netOpbyte(0x86) // Event for Trace Exception (Memory exception)
)

// 68000 has pins called UDS(Upper Data Strobe) and LDS(Lower Data Strobe), and these signals tell the bus to only look at upper or lower 8-bit of the 16-bit external bus.
// This is to allow writing/reading 8-bit values without touching the other half.
type netDs uint8

const (
	netDsUpper = netDs(1 << 0)
	netDsLower = netDs(1 << 1)
	netDsBoth  = netDs(netDsLower | netDsUpper)
)

func (ctx *clientContext) main() {
	logger := log.New(log.Writer(), fmt.Sprintf("[client/%s] ", ctx.conn.RemoteAddr()), log.Flags())
	for !ctx.closed {
		err := ctx.serveNextCmd(logger)
		if err != nil {
			logger.Printf("Closing client connection due to an error: %v", err)
			break
		}
	}
	logger.Printf("Closing client connection")
	ctx.conn.Close()
	logger.Printf("Closed client connection")
}
func (ctx *clientContext) serveNextCmd(logger *log.Logger) error {
	const (
		debugNetmsg = false
	)

	var hdrByte uint8
	hdrByte, err := ctx.inB()
	if err != nil {
		return err
	}
	switch netOpbyte(hdrByte) {
	case netOpbyteBye:
		if debugNetmsg {
			logger.Printf("Bye")
		}
		ctx.closed = true

	case netOpbyteUnstop:
		if debugNetmsg {
			logger.Printf("Unstop")
		}
		ctx.stopped = false
		res := newNetAckResponse(0)
		if err := ctx.out(res); err != nil {
			return err
		}

	case netOpbyteIsStopped:
		if debugNetmsg {
			logger.Printf("IsStopped")
		}
		res := newNetAckResponse(1)
		if ctx.stopped {
			res.appendB(1)
		} else {
			res.appendB(0)
		}
		if err := ctx.out(res); err != nil {
			return err
		}

	case netOpbyteTraceExecOn:
		if debugNetmsg {
			logger.Printf("TraceExecOn")
		}
		ctx.traceExec = true
		res := newNetAckResponse(0)
		if err := ctx.out(res); err != nil {
			return err
		}

	case netOpbyteTraceExecOff:
		if debugNetmsg {
			logger.Printf("TraceExecOff")
		}
		ctx.traceExec = false
		res := newNetAckResponse(0)
		if err := ctx.out(res); err != nil {
			return err
		}

	case netOpbyteTraceExcOn:
		if debugNetmsg {
			logger.Printf("TraceExcOn")
		}
		ctx.traceExc = true
		res := newNetAckResponse(0)
		if err := ctx.out(res); err != nil {
			return err
		}

	case netOpbyteTraceExcOff:
		if debugNetmsg {
			logger.Printf("TraceExcOff")
		}
		ctx.traceExc = false
		res := newNetAckResponse(0)
		if err := ctx.out(res); err != nil {
			return err
		}

	case netOpbyteTick:
		if debugNetmsg {
			logger.Printf("Tick")
		}
		if ctx.stopped {
			res := newNetAckResponse(0)
			if err := ctx.out(res); err != nil {
				return err
			}
		} else {
			instrPc := ctx.pc
			ctx.decodingCtx = decodingContext{}
			executed := false
			err := func() error {
				if v, err := ctx.fetchInstrW(); err != nil {
					return err
				} else {
					ctx.decodingCtx.ir = v
				}
				if (ctx.decodingCtx.ir >> 12) == 0xa {
					return excError{exc: excLineA}
				}
				if (ctx.decodingCtx.ir >> 12) == 0xf {
					return excError{exc: excLineF}
				}
				instr, err := ctx.instrDecode()
				if err != nil {
					return err
				}
				if ctx.traceExec {
					disasm := instr.disasm()
					if err := ctx.eventTraceExec(instrPc, ctx.decodingCtx.ir, disasm); err != nil {
						return err
					}
				}
				if err := instr.exec(ctx); err != nil {
					if excErr, isExcErr := err.(excError); isExcErr && excErr.exc == excPrivilegeViolation {
						executed = false
					} else {
						executed = true
					}
					return err
				}
				executed = true
				return nil
			}()
			if !executed {
				ctx.pc = instrPc
			} else {
				ctx.lastExecutedIr = ctx.decodingCtx.ir
			}
			if err != nil {
				if excErr, isExcErr := err.(excError); isExcErr {
					res := newNetAckResponse(0)
					if err := ctx.beginExc(excErr); err != nil {
						logger.Printf("beginExc failed with error: %v", err)
						res = newNetFailResponse()
					}
					if err := ctx.out(res); err != nil {
						return err
					}
				} else {
					// Non-exception error occured
					return err
				}
			} else {
				res := newNetAckResponse(0)
				if err := ctx.out(res); err != nil {
					return err
				}
			}
		}

	case netOpbyteDregWrite:
		reg, err := ctx.inB()
		if err != nil {
			return err
		}
		val, err := ctx.inL()
		if err != nil {
			return err
		}
		if debugNetmsg {
			logger.Printf("DregWrite %d %#x", reg, val)
		}
		if 7 < reg {
			return ctx.outFail()
		}
		ctx.writeDregL(reg, val)
		res := newNetAckResponse(0)
		if err := ctx.out(res); err != nil {
			return err
		}

	case netOpbyteDregRead:
		reg, err := ctx.inB()
		if err != nil {
			return err
		}
		if debugNetmsg {
			logger.Printf("DregRead %d", reg)
		}
		if 7 < reg {
			return ctx.outFail()
		}
		val := ctx.readDregL(reg)
		res := newNetAckResponse(4)
		res.appendL(val)
		if err := ctx.out(res); err != nil {
			return err
		}

	case netOpbyteAregWrite:
		reg, err := ctx.inB()
		if err != nil {
			return err
		}
		val, err := ctx.inL()
		if err != nil {
			return err
		}
		if debugNetmsg {
			logger.Printf("AregWrite %d %#x", reg, val)
		}
		if 7 < reg {
			return ctx.outFail()
		}
		ctx.writeAregL(reg, val)
		res := newNetAckResponse(0)
		if err := ctx.out(res); err != nil {
			return err
		}

	case netOpbyteAregRead:
		reg, err := ctx.inB()
		if err != nil {
			return err
		}
		if debugNetmsg {
			logger.Printf("AregRead %d", reg)
		}
		if 7 < reg {
			return ctx.outFail()
		}
		val := ctx.readAreg(reg)
		res := newNetAckResponse(4)
		res.appendL(val)
		if err := ctx.out(res); err != nil {
			return err
		}

	case netOpbyteSspWrite:
		val, err := ctx.inL()
		if err != nil {
			return err
		}
		if debugNetmsg {
			logger.Printf("SspWrite %#x", val)
		}
		ctx.a7ssp = val
		res := newNetAckResponse(0)
		if err := ctx.out(res); err != nil {
			return err
		}

	case netOpbyteSspRead:
		if debugNetmsg {
			logger.Printf("SspRead")
		}
		res := newNetAckResponse(4)
		res.appendL(ctx.a7ssp)
		if err := ctx.out(res); err != nil {
			return err
		}

	case netOpbyteUspWrite:
		val, err := ctx.inL()
		if err != nil {
			return err
		}
		if debugNetmsg {
			logger.Printf("UspWrite %#x", val)
		}
		ctx.a7usp = val
		res := newNetAckResponse(0)
		if err := ctx.out(res); err != nil {
			return err
		}

	case netOpbyteUspRead:
		if debugNetmsg {
			logger.Printf("UspRead")
		}
		res := newNetAckResponse(4)
		res.appendL(ctx.a7usp)
		if err := ctx.out(res); err != nil {
			return err
		}

	case netOpbytePcWrite:
		val, err := ctx.inL()
		if err != nil {
			return err
		}
		if debugNetmsg {
			logger.Printf("PcWrite %#x", val)
		}
		ctx.pc = val
		res := newNetAckResponse(0)
		if err := ctx.out(res); err != nil {
			return err
		}

	case netOpbytePcRead:
		if debugNetmsg {
			logger.Printf("PcRead")
		}
		res := newNetAckResponse(4)
		res.appendL(ctx.pc)

		if err := ctx.out(res); err != nil {
			return err
		}

	case netOpbyteSrWrite:
		sr, err := ctx.inW()
		if err != nil {
			return err
		}
		if debugNetmsg {
			logger.Printf("SrWrite %#x", sr)
		}
		ctx.writeSr(sr)
		res := newNetAckResponse(0)
		if err := ctx.out(res); err != nil {
			return err
		}

	case netOpbyteSrRead:
		if debugNetmsg {
			logger.Printf("SrRead")
		}
		res := newNetAckResponse(2)
		res.appendW(ctx.readSr())
		if err := ctx.out(res); err != nil {
			return err
		}

	default:
		logger.Printf("Unrecognized message type %x", hdrByte)
		if err := ctx.outFail(); err != nil {
			return err
		}
	}
	return nil
}

func (ctx *clientContext) eventAddrAsserted(addr uint32) error {
	// Send event --------------------------------------------------------------
	event := newNetEvent(netOpbyteEventAddrAsserted, 4)
	event.appendL(addr)
	if err := ctx.out(event); err != nil {
		return err
	}
	// Receive response --------------------------------------------------------
	return ctx.expectAckOrFail()
}
func (ctx *clientContext) eventReadBus(ds netDs) (uint16, error) {
	// Send event --------------------------------------------------------------
	event := newNetEvent(netOpbyteEventReadBus, 1)
	event.appendB(uint8(ds))
	if err := ctx.out(event); err != nil {
		return 0, err
	}
	// Receive response --------------------------------------------------------
	if err := ctx.expectAckOrFail(); err != nil {
		return 0, err
	}
	return ctx.inW()
}
func (ctx *clientContext) eventWriteBus(ds netDs, v uint16) error {
	// Send event --------------------------------------------------------------
	event := newNetEvent(netOpbyteEventWriteBus, 3)
	event.appendB(uint8(ds))
	event.appendW(v)
	if err := ctx.out(event); err != nil {
		return err
	}
	// Receive response --------------------------------------------------------
	return ctx.expectAckOrFail()
}
func (ctx *clientContext) eventTraceExec(pc uint32, ir uint16, disasm string) error {
	// Send event --------------------------------------------------------------
	event := newNetEvent(netOpbyteEventTraceExec, 7+len(disasm))
	event.appendL(pc)
	event.appendW(ir)
	event.appendS(disasm)
	if err := ctx.out(event); err != nil {
		return err
	}
	// Receive response --------------------------------------------------------
	return ctx.expectAckOrFail()
}
func (ctx *clientContext) eventReset() error {
	// Send event --------------------------------------------------------------
	event := newNetEvent(netOpbyteEventReset, 0)
	if err := ctx.out(event); err != nil {
		return err
	}
	// Receive response --------------------------------------------------------
	return ctx.expectAckOrFail()
}
func (ctx *clientContext) eventTraceExc(exc exc, pc uint32) error {
	// Send event --------------------------------------------------------------
	event := newNetEvent(netOpbyteEventTraceExc, 5)
	event.appendB(uint8(exc))
	event.appendL(pc)
	if err := ctx.out(event); err != nil {
		return err
	}
	// Receive response --------------------------------------------------------
	return ctx.expectAckOrFail()
}
func (ctx *clientContext) eventTraceExcMem(exc exc, pc uint32, ir uint16, errAddr uint32, flags uint8) error {
	// Send event --------------------------------------------------------------
	event := newNetEvent(netOpbyteEventTraceExcMem, 12)
	event.appendB(uint8(exc))
	event.appendL(pc)
	event.appendW(ir)
	event.appendL(errAddr)
	event.appendB(flags)
	if err := ctx.out(event); err != nil {
		return err
	}
	// Receive response --------------------------------------------------------
	return ctx.expectAckOrFail()
}

type sendBuf struct {
	buf  []uint8
	dest []uint8
}

func newNetEvent(typ netOpbyte, restLen int) sendBuf {
	buf := make([]uint8, restLen+1)
	buf[0] = uint8(typ)
	return sendBuf{buf: buf, dest: buf[1:]}
}
func newNetAckResponse(restLen int) sendBuf {
	buf := make([]uint8, restLen+1)
	buf[0] = uint8(netOpbyteAck)
	return sendBuf{buf: buf, dest: buf[1:]}
}
func newNetFailResponse() sendBuf {
	buf := make([]uint8, 1)
	buf[0] = uint8(netOpbyteFail)
	return sendBuf{buf: buf, dest: buf[1:]}
}

func (b *sendBuf) appendB(v uint8) {
	b.dest[0] = v
	b.dest = b.dest[1:]
}
func (b *sendBuf) appendW(v uint16) {
	binary.BigEndian.PutUint16(b.dest[0:2], v)
	b.dest = b.dest[2:]
}
func (b *sendBuf) appendL(v uint32) {
	binary.BigEndian.PutUint32(b.dest[0:4], v)
	b.dest = b.dest[4:]
}
func (b *sendBuf) appendS(s string) {
	if 255 < len(s) {
		panic("string cannot be sent because it's too long(max: 255 bytes)")
	}
	b.appendB(byte(len(s)))
	for i := 0; i < len(s); i++ {
		b.dest[0] = s[i]
		b.dest = b.dest[1:]
	}
}

func (ctx *clientContext) out(b sendBuf) error {
	// Make sure we were not wasting more space by accident
	if len(b.dest) != 0 {
		panic("too many bytes were allocated")
	}
	_, err := ctx.conn.Write(b.buf)
	return err
}
func (ctx *clientContext) outFail() error {
	return ctx.out(newNetFailResponse())
}

func (ctx *clientContext) inB() (uint8, error) {
	return ctx.reader.ReadByte()
}
func (ctx *clientContext) inW() (uint16, error) {
	bytes := [2]uint8{}
	_, err := io.ReadFull(ctx.reader, bytes[:])
	if err != nil {
		return 0, err
	}
	res := (uint16(bytes[0]) << 8) | uint16(bytes[1])
	return res, nil
}
func (ctx *clientContext) inL() (uint32, error) {
	bytes := [4]uint8{}
	_, err := io.ReadFull(ctx.reader, bytes[:])
	if err != nil {
		return 0, err
	}
	res := (uint32(bytes[0]) << 24) | (uint32(bytes[1]) << 16) | (uint32(bytes[2]) << 8) | uint32(bytes[3])
	return res, nil
}
func (ctx *clientContext) expectAckOrFail() error {
	ackByte, err := ctx.inB()
	if err != nil {
		return err
	}
	switch netOpbyte(ackByte) {
	case netOpbyteAck:
		return nil
	case netOpbyteFail:
		return fmt.Errorf("communication error: expected ACK(%#x) got FAIL(%#x)", netOpbyteAck, netOpbyteFail)
	default:
		return fmt.Errorf("communication error: expected ACK(%#x) or FAIL(%#x), got %#x", netOpbyteAck, netOpbyteFail, ackByte)
	}
}

//==============================================================================
// Misc utilities
//==============================================================================

type regType uint8

const (
	regTypeAddr = regType(iota)
	regTypeData
)

func (t regType) ToString() string {
	switch t {
	case regTypeAddr:
		return "a"
	case regTypeData:
		return "d"
	default:
		log.Panicf("unrecognized regType value %d", t)
		return "???"
	}
}

func signExtendWToL(v uint16) uint32 {
	return uint32(int32(int16(v)))
}
func signExtendBToW(v uint8) uint16 {
	return uint16(int16(int8(v)))
}
func signExtendBToL(v uint8) uint32 {
	return uint32(int32(int8(v)))
}

//==============================================================================
// Below are instruction implementations
//==============================================================================

// ==============================================================================
// Instructions: Data movement
// ==============================================================================

// MOVE.b
func (instr instrMoveB) disasm() string {
	return fmt.Sprintf("move.b %s %s", instr.ea1.ToString(), instr.ea2.ToString())
}
func (instr instrMoveB) exec(ctx *clientContext) error {
	src := uint8(0)
	if v, err := ctx.readEa(*instr.ea1, opsizeByte); err != nil {
		return err
	} else {
		src = uint8(v)
	}
	if err := ctx.writeEa(*instr.ea2, opsizeByte, uint32(src)); err != nil {
		return err
	}
	ctx.setNZFlagsB(src)
	ctx.clearVCFlags()
	return nil
}

// ==============================================================================
// Instructions: Branching
//
// Note that PC value in branching instructions are always <address of instruction word> + 2.
// this includes the reported PC when address error occurs.
// ==============================================================================

// BRA
func (instr instrBra) disasm() string {
	addr := instr.instrPc + 2 + instr.branchOff
	return fmt.Sprintf("bra %#x", addr)
}
func (instr instrBra) exec(ctx *clientContext) error {
	addr := instr.instrPc + 2 + instr.branchOff
	if (addr & 0x1) != 0 {
		ctx.pc = instr.instrPc + 2
		return ctx.memExcError(excAddressError, addr, ctx.getFuncCode(true), busDirRead)
	}
	ctx.pc = addr
	return nil
}

// BSR
func (instr instrBsr) disasm() string {
	addr := instr.instrPc + 2 + instr.branchOff
	return fmt.Sprintf("bsr %#x", addr)
}
func (instr instrBsr) exec(ctx *clientContext) error {
	if err := ctx.pushL(ctx.pc); err != nil {
		return err
	}
	addr := instr.instrPc + 2 + instr.branchOff
	if (addr & 0x1) != 0 {
		// BSR is not like other instructions
		// when address error occurs, reported PC is at the new address
		ctx.pc = addr
		return ctx.memExcError(excAddressError, addr, ctx.getFuncCode(true), busDirRead)
	}
	ctx.pc = addr
	return nil
}

// Bcc
func (instr instrBcc) disasm() string {
	addr := instr.instrPc + 2 + instr.branchOff
	return fmt.Sprintf("b%s %#x", instr.cond.ToString(), addr)
}
func (instr instrBcc) exec(ctx *clientContext) error {
	if !ctx.testCond(instr.cond) {
		return nil
	}
	addr := instr.instrPc + 2 + instr.branchOff
	if (addr & 0x1) != 0 {
		ctx.pc = instr.instrPc + 2
		return ctx.memExcError(excAddressError, addr, ctx.getFuncCode(true), busDirRead)
	}
	ctx.pc = addr
	return nil
}

// DBcc
func (instr instrDbcc) disasm() string {
	addr := instr.instrPc + 2 + signExtendWToL(instr.imm16)
	return fmt.Sprintf("db%s D%d %#x", instr.cond.ToString(), instr.regY, addr)
}
func (instr instrDbcc) exec(ctx *clientContext) error {
	if ctx.testCond(instr.cond) {
		return nil
	}
	addr := instr.instrPc + 2 + signExtendWToL(instr.imm16)
	if (addr & 0x1) != 0 {
		return ctx.memExcError(excAddressError, addr, ctx.getFuncCode(true), busDirRead)
	}
	dn := ctx.readDregW(instr.regY)
	dn -= 1
	ctx.writeDregW(instr.regY, dn)
	if dn == 0xffff {
		return nil
	}
	ctx.pc = addr
	return nil
}

// ==============================================================================
// Instructions: Return series
// ==============================================================================

func (instr instrRts) disasm() string {
	return "rts"
}
func (instr instrRts) exec(ctx *clientContext) error {
	if v, err := ctx.popL(); err != nil {
		return err
	} else if (v & 0x1) != 0 {
		return ctx.memExcError(excAddressError, v, ctx.getFuncCode(true), busDirRead)
	} else {
		ctx.pc = v
	}
	return nil
}

func (instr instrRtr) disasm() string {
	return "rtr"
}
func (instr instrRtr) exec(ctx *clientContext) error {
	if v, err := ctx.popW(); err != nil {
		return err
	} else {
		ctx.writeCcr(uint8(v))
	}
	if v, err := ctx.popL(); err != nil {
		return err
	} else if (v & 0x1) != 0 {
		return ctx.memExcError(excAddressError, v, ctx.getFuncCode(true), busDirRead)
	} else {
		ctx.pc = v
	}
	return nil
}

func (instr instrRte) disasm() string {
	return "rte"
}
func (instr instrRte) exec(ctx *clientContext) error {
	if !ctx.srS {
		return excError{exc: excPrivilegeViolation}
	}
	newSr := uint16(0)
	if v, err := ctx.popW(); err != nil {
		return err
	} else {
		// Note that we don't update SR yet, so that we don't switch to USP stack before we are done.
		newSr = v
	}
	if v, err := ctx.popL(); err != nil {
		return err
	} else if (v & 0x1) != 0 {
		ctx.writeSr(newSr)
		return ctx.memExcError(excAddressError, v, ctx.getFuncCode(true), busDirRead)
	} else {
		ctx.writeSr(newSr)
		ctx.pc = v
	}
	return nil
}

// ==============================================================================
// Instructions: Misc
// ==============================================================================

// LEA
func (instr instrLea) disasm() string {
	return fmt.Sprintf("lea %s, a%d", instr.ea1.ToString(), instr.regX)
}
func (instr instrLea) exec(ctx *clientContext) error {
	addr := ctx.memAddrOfEa(*instr.ea1, opsizeNone)
	ctx.writeAregL(instr.regX, addr)
	return nil
}

// PEA
func (instr instrPea) disasm() string {
	return fmt.Sprintf("pea %s", instr.ea1.ToString())
}
func (instr instrPea) exec(ctx *clientContext) error {
	addr := ctx.memAddrOfEa(*instr.ea1, opsizeNone)
	return ctx.pushL(addr)
}

// JMP
func (instr instrJmp) disasm() string {
	return fmt.Sprintf("jmp %s", instr.ea1.ToString())
}
func (instr instrJmp) exec(ctx *clientContext) error {
	addr := ctx.memAddrOfEa(*instr.ea1, opsizeNone)
	if (addr & 0x1) != 0 {
		// Address error during JMP always seem to push (instruction address + 2),
		// regardless of addressing mode.
		ctx.pc = instr.instrPc + 2
		return ctx.memExcError(excAddressError, addr, ctx.getFuncCode(true), busDirRead)
	}
	ctx.pc = addr
	return nil
}

// JSR
func (instr instrJsr) disasm() string {
	return fmt.Sprintf("jsr %s", instr.ea1.ToString())
}
func (instr instrJsr) exec(ctx *clientContext) error {
	addr := ctx.memAddrOfEa(*instr.ea1, opsizeNone)
	if (addr & 0x1) != 0 {
		return ctx.memExcError(excAddressError, addr, ctx.getFuncCode(true), busDirRead)
	}
	if err := ctx.pushL(ctx.pc); err != nil {
		return err
	}
	ctx.pc = addr
	return nil
}

// LINK
func (instr instrLink) disasm() string {
	return fmt.Sprintf("link a%d, #%d", instr.regY, instr.imm16)
}
func (instr instrLink) exec(ctx *clientContext) error {
	addr := ctx.readAreg(instr.regY)
	if err := ctx.pushL(addr); err != nil {
		return err
	}
	sp := ctx.readAreg(7)
	ctx.writeAregL(instr.regY, sp)
	sp += signExtendWToL(instr.imm16)
	ctx.writeAregL(7, sp)
	return nil
}

// UNLK
func (instr instrUnlk) disasm() string {
	return fmt.Sprintf("unlk a%d", instr.regY)
}
func (instr instrUnlk) exec(ctx *clientContext) error {
	sp := ctx.readAreg(instr.regY)
	if (sp & 0x1) != 0 {
		// XXX: For some reason, UNLK test expects PC+2 instead of PC? huh...?
		ctx.pc += 2
		return ctx.memExcError(excAddressError, sp, ctx.getFuncCode(false), busDirRead)
	}
	ctx.writeAregL(7, sp)
	if v, err := ctx.popL(); err != nil {
		return err
	} else {
		ctx.writeAregL(instr.regY, v)
	}
	return nil
}

// TRAP
func (instr instrTrap) disasm() string {
	return fmt.Sprintf("trap #%d", instr.vector)
}
func (instr instrTrap) exec(ctx *clientContext) error {
	return ctx.beginExc(excError{exc: excTrapVectorStart + exc(instr.vector)})
}

// TRAPV
func (instr instrTrapV) disasm() string {
	return "trapv"
}
func (instr instrTrapV) exec(ctx *clientContext) error {
	if !ctx.ccrV {
		return nil
	}
	return ctx.beginExc(excError{exc: excTrapv})
}

// EXT.w
func (instr instrExtW) disasm() string {
	return fmt.Sprintf("ext.w d%d", instr.regY)
}
func (instr instrExtW) exec(ctx *clientContext) error {
	val8 := ctx.readDregB(instr.regY)
	val16 := signExtendBToW(val8)
	ctx.writeDregW(instr.regY, val16)
	ctx.setNZFlagsW(val16)
	ctx.clearVCFlags()
	return nil
}

// EXT.l
func (instr instrExtL) disasm() string {
	return fmt.Sprintf("ext.l d%d", instr.regY)
}
func (instr instrExtL) exec(ctx *clientContext) error {
	val16 := ctx.readDregW(instr.regY)
	val32 := signExtendWToL(val16)
	ctx.writeDregL(instr.regY, val32)
	ctx.setNZFlagsL(val32)
	ctx.clearVCFlags()
	return nil
}

// MOVE An, USP
func (instr instrMoveToUsp) disasm() string {
	return fmt.Sprintf("move a%d, usp", instr.regY)
}
func (instr instrMoveToUsp) exec(ctx *clientContext) error {
	if !ctx.srS {
		return excError{exc: excPrivilegeViolation}
	}
	ctx.a7usp = ctx.readAreg(instr.regY)
	return nil
}

// MOVE USP, An
func (instr instrMoveFromUsp) disasm() string {
	return fmt.Sprintf("move usp, a%d", instr.regY)
}
func (instr instrMoveFromUsp) exec(ctx *clientContext) error {
	if !ctx.srS {
		return excError{exc: excPrivilegeViolation}
	}
	ctx.writeAregL(instr.regY, ctx.a7usp)
	return nil
}

// EXG Dn,Dn
func (instr instrExgDReg) disasm() string {
	return fmt.Sprintf("exg d%d, d%d", instr.regY, instr.regX)
}
func (instr instrExgDReg) exec(ctx *clientContext) error {
	x := ctx.readDregL(instr.regX)
	y := ctx.readDregL(instr.regY)
	ctx.writeDregL(instr.regX, y)
	ctx.writeDregL(instr.regY, x)
	return nil
}

// EXG An,An
func (instr instrExgAReg) disasm() string {
	return fmt.Sprintf("exg a%d, a%d", instr.regY, instr.regX)
}
func (instr instrExgAReg) exec(ctx *clientContext) error {
	x := ctx.readAreg(instr.regX)
	y := ctx.readAreg(instr.regY)
	ctx.writeAregL(instr.regX, y)
	ctx.writeAregL(instr.regY, x)
	return nil
}

// EXG Dn,An
func (instr instrExgDAReg) disasm() string {
	return fmt.Sprintf("exg d%d, a%d", instr.regY, instr.regX)
}
func (instr instrExgDAReg) exec(ctx *clientContext) error {
	x := ctx.readDregL(instr.regX)
	y := ctx.readAreg(instr.regY)
	ctx.writeDregL(instr.regX, y)
	ctx.writeAregL(instr.regY, x)
	return nil
}

// SWAP
func (instr instrSwap) disasm() string {
	return fmt.Sprintf("swap d%d", instr.regY)
}
func (instr instrSwap) exec(ctx *clientContext) error {
	res := ctx.readDregL(instr.regY)
	res = ((res & 0xffff0000) >> 16) | ((res & 0x0000ffff) << 16)
	ctx.setNZFlagsL(res)
	ctx.clearVCFlags()
	ctx.writeDregL(instr.regY, res)
	return nil
}

// ILLEGAL
func (instr instrIllegal) disasm() string {
	return "illegal"
}
func (instr instrIllegal) exec(ctx *clientContext) error {
	return excError{exc: excIllegalInstr}
}

// NOP
func (instr instrNop) disasm() string {
	return "nop"
}
func (instr instrNop) exec(ctx *clientContext) error {
	return nil
}

// RESET
func (instr instrReset) disasm() string {
	return "reset"
}
func (instr instrReset) exec(ctx *clientContext) error {
	if !ctx.srS {
		return excError{exc: excPrivilegeViolation}
	}
	return ctx.eventReset()
}
