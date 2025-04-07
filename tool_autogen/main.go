// Copyright (c) 2025, Oh Inseo (YJK) -- Licensed under BSD-2-Clause
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

// Note that spaces in the instruction name are ignored.
// They are only there for the source code formatting.
var records []record = []record{
	// Branch ------------------------------------------------------------------
	{"Bra ", "01100000bbbbbbbb", []*field{}, xwordBranchOff, eamodeFlagNone, eamodeFlagNone},
	{"Bsr ", "01100001bbbbbbbb", []*field{}, xwordBranchOff, eamodeFlagNone, eamodeFlagNone},
	{"Bcc ", "0110aaaabbbbbbbb", []*field{fieldCond}, xwordBranchOff, eamodeFlagNone, eamodeFlagNone},
	{"Dbcc", "0101aaaa11001bbb", []*field{fieldCond, fieldRegY}, xwordImm16, eamodeFlagNone, eamodeFlagNone},

	// Misc(0100~) -------------------------------------------------------------
	{"Lea        ", "0100aaa111bbbbbb", []*field{fieldRegX, fieldEa1}, nil, eamodeFlagCtrl, eamodeFlagNone},
	{"Pea        ", "0100100001aaaaaa", []*field{fieldEa1}, nil, eamodeFlagCtrl, eamodeFlagNone},
	{"Jmp        ", "0100111011aaaaaa", []*field{fieldEa1}, nil, eamodeFlagCtrl, eamodeFlagNone},
	{"Jsr        ", "0100111010aaaaaa", []*field{fieldEa1}, nil, eamodeFlagCtrl, eamodeFlagNone},
	{"Link       ", "0100111001010aaa", []*field{fieldRegY}, xwordImm16, eamodeFlagNone, eamodeFlagNone},
	{"Unlk       ", "0100111001011aaa", []*field{fieldRegY}, nil, eamodeFlagNone, eamodeFlagNone},
	{"Swap       ", "0100100001000aaa", []*field{fieldRegY}, nil, eamodeFlagNone, eamodeFlagNone},
	{"MoveToUsp  ", "0100111001100aaa", []*field{fieldRegY}, nil, eamodeFlagNone, eamodeFlagNone},
	{"MoveFromUsp", "0100111001101aaa", []*field{fieldRegY}, nil, eamodeFlagNone, eamodeFlagNone},
	{"ExtW       ", "0100100010000bbb", []*field{fieldRegY}, nil, eamodeFlagNone, eamodeFlagNone},
	{"ExtL       ", "0100100011000bbb", []*field{fieldRegY}, nil, eamodeFlagNone, eamodeFlagNone},
	{"Trap       ", "010011100100aaaa", []*field{fieldVector}, nil, eamodeFlagNone, eamodeFlagNone},
	{"TrapV      ", "0100111001110110", []*field{}, nil, eamodeFlagNone, eamodeFlagNone},

	{"ExgDReg    ", "1100aaa101000ccc", []*field{fieldRegX, fieldRegY}, nil, eamodeFlagNone, eamodeFlagNone},
	{"ExgAReg    ", "1100aaa101001ccc", []*field{fieldRegX, fieldRegY}, nil, eamodeFlagNone, eamodeFlagNone},
	{"ExgDAReg   ", "1100aaa110001ccc", []*field{fieldRegX, fieldRegY}, nil, eamodeFlagNone, eamodeFlagNone},

	// Misc(Without any fields) ------------------------------------------------
	{"Illegal", "0100101011111100", []*field{}, nil, eamodeFlagNone, eamodeFlagNone},
	{"Nop    ", "0100111001110001", []*field{}, nil, eamodeFlagNone, eamodeFlagNone},
	{"Rts    ", "0100111001110101", []*field{}, nil, eamodeFlagNone, eamodeFlagNone},
	{"Rtr    ", "0100111001110111", []*field{}, nil, eamodeFlagNone, eamodeFlagNone},
	{"Reset  ", "0100111001110000", []*field{}, nil, eamodeFlagNone, eamodeFlagNone},
	{"Rte    ", "0100111001110011", []*field{}, nil, eamodeFlagNone, eamodeFlagNone},
}

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Printf("Usage: %s <output path>\n", args[0])
		os.Exit(1)
	}
	fileName := args[1]
	outFile, err := os.Create(fileName)
	if err != nil {
		fmt.Printf("Could not open file %s\n", fileName)
		os.Exit(1)
	}
	defer outFile.Close()

	dest = bufio.NewWriter(outFile)
	defer dest.Flush()

	date := time.Now().Format(time.DateTime)
	emitln("// This file was automatically generated.")
	emitln("// Generated at %s", date)
	emitln("package main")
	emitln("")
	//==========================================================================
	// Output the instruction struct and function prototypes
	//==========================================================================
	for _, rec := range records {
		fmt.Println("Generating struct for:", rec)
		emitBeginBlock("type %s struct", rec.structName())
		{
			emitln("instrPc uint32")
			emitln("")

			// Output fields ---------------------------------------------------
			for _, field := range rec.fields {
				fmt.Println(" - Field:", field)
				emitln("%s %s", field.fieldName, field.typeName)
			}
			emitln("")

			// Output extension words ------------------------------------------
			if rec.xword != nil {
				fmt.Println(" - Xword:", rec.xword)
				emitln("%s %s", rec.xword.fieldName, rec.xword.typeName)
			}
		}

		emitEndBlock("")
		emitln("")
	}
	//==========================================================================
	// Output the decoder function
	//==========================================================================
	emitln("//==========================================================================")
	emitln("// Decoder function")
	emitln("//==========================================================================")
	emitBeginBlock("func (ctx *clientContext) instrDecode() (res instr, err error)")
	for _, rec := range records {
		fmt.Println("Generating decoder for:", rec)

		// Parse the bit string --------------------------------------------
		mask := uint16(1 << 15)
		fixedMask := uint16(0)
		fixedValue := uint16(0)
		if len(rec.bits) != 16 {
			fmt.Printf("[%s] bit string (%s) must be 16 characters\n", rec.structName(), rec.bits)
			continue
		}
		for i := range 16 {
			c := rec.bits[i]
			if (c == '0') || (c == '1') {
				fixedMask |= mask
			}
			if c == '1' {
				fixedValue |= mask
			}
			mask >>= 1
		}

		emitln("// %s", rec.structName())
		emitBeginBlock("func()")
		{
			emitln("err = nil")
			emitln("resTemp := %s{}", rec.structName())
			emitln("resTemp.instrPc = ctx.pc - 2")

			// Check the bit pattern -------------------------------------------
			emitBeginBlock("if (ctx.decodingCtx.ir & %#x) != %#x", fixedMask, fixedValue)
			{
				emitln("err = excError{exc: excIllegalInstr}")
				emitln("return")
			}
			emitEndBlock("")

			// Output field decoder calls --------------------------------------
			for _, field := range rec.fields {
				fmt.Println(" - Field:", field)
				emitBeginBlock("if v, ok := ctx.decodeField%s(); !ok", field.decoderName)
				{
					emitln("err = excError{exc: excIllegalInstr}")
					emitln("return")
				}
				emitEndBeginBlock("else")
				{
					emitln("resTemp.%s = v", field.fieldName)
				}
				emitEndBlock("")
			}

			// Check EA mode ---------------------------------------------------
			makeEaArrayCode := func(modes eamodeFlag) string {
				sb := strings.Builder{}
				sb.WriteString("[]eamode{")
				for n, mode := range modes.getConstNames() {
					if n != 0 {
						sb.WriteString(", ")
					}
					sb.WriteString(mode)
				}
				sb.WriteString("}")
				return sb.String()
			}
			emitBeginBlock("if !ctx.checkEaModes(%s, %s)", makeEaArrayCode(rec.ea1), makeEaArrayCode(rec.ea2))
			{
				emitln("err = excError{exc: excIllegalInstr}")
				emitln("return")
			}
			emitEndBlock("")

			// Call extension word decoder -------------------------------------
			if rec.xword != nil {
				fmt.Println(" - Xword:", rec.xword)
				emitBeginBlock("if v, err := ctx.decodeXword%s(); err != nil", rec.xword.decoderName)
				{
					emitln("err = excError{exc: excIllegalInstr}")
					emitln("return")
				}
				emitEndBeginBlock("else")
				{
					emitln("resTemp.%s = v", rec.xword.fieldName)
				}
				emitEndBlock("")
			}

			// Call EA decoder -------------------------------------------------
			emitBeginBlock("if err = ctx.decodeEa(); err != nil")
			{
				emitln("return")
			}
			emitEndBlock("")

			// Success ---------------------------------------------------------
			emitln("res = resTemp")
		}
		emitEndBlock("()")
		// Check decoding result -----------------------------------------------
		emitBeginBlock("if excErr, isExcErr := err.(excError); !isExcErr || (isExcErr && (excErr.exc != excIllegalInstr))")
		{
			emitln("return")
		}
		emitEndBlock("")
	}

	emitln("return nil, excError{exc: excIllegalInstr}")
	emitEndBlock("")

	fmt.Println("hello, world")
}

//==============================================================================
// 68000 instructions
//==============================================================================

type record struct {
	name   string     // Instruction name
	bits   string     // Instruction bit pattern. 0/1 are fixed bits, others are for fields
	fields []*field   // Instruction fields
	xword  *xword     // Extension word type (nil if not present)
	ea1    eamodeFlag // Effective address mode 1
	ea2    eamodeFlag // Effective address mode 2 (if present)
}

func (rec record) structName() string {
	return fmt.Sprintf("instr%s", strings.TrimSpace(rec.name))
}

// Instruction fields
// type field uint8
type field struct {
	typeName    string
	fieldName   string
	decoderName string
}

var (
	fieldCond   *field = &field{"cond", "cond", "Cond"}
	fieldEa1    *field = &field{"*ea", "ea1", "Ea1"}
	fieldEa2    *field = &field{"*ea", "ea2", "Ea2"}
	fieldRegX   *field = &field{"uint8", "regX", "RegX"}
	fieldRegY   *field = &field{"uint8", "regY", "RegY"}
	fieldImm3   *field = &field{"uint8", "imm", "Imm3"}
	fieldImm8   *field = &field{"uint8", "imm", "Imm8"}
	fieldVector *field = &field{"uint8", "vector", "Vector"}
)

// Instruction extension word
type xword struct {
	typeName    string
	fieldName   string
	decoderName string
}

var (
	xwordBranchOff *xword = &xword{"uint32", "branchOff", "BranchOff"} // Branch offset
	xwordImm       *xword = &xword{"uint32", "imm", "Imm"}             // 8/16/32-bit immediate data (Determined based on operation size)
	xwordImm8      *xword = &xword{"uint8", "imm8", "Imm8"}            // 8-bit immediate data
	xwordImm16     *xword = &xword{"uint16", "imm16", "Imm16"}         // 16-bit immediate data
)

const ()

// Effective addressing modes
type eamodeFlag uint16

const (
	eamodeFlagNone = 0
	eamodeFlagAll  = eamodeFlagData | eamodeFlagMem | eamodeFlagCtrl | eamodeFlagAlter

	eamodeFlagDreg           = eamodeFlag(1 << 0)  // Data register direct [Dn]
	eamodeFlagAreg           = eamodeFlag(1 << 1)  // Address register direct [An]
	eamodeFlagAregInd        = eamodeFlag(1 << 2)  // Address register indirect [(An)]
	eamodeFlagAregIndPostinc = eamodeFlag(1 << 3)  // Address register indirect + postincrement [(An)+]
	eamodeFlagAregIndPredec  = eamodeFlag(1 << 4)  // Address register indirect + predecrement [-(An)]
	eamodeFlagAregIndDisp    = eamodeFlag(1 << 5)  // Address register indirect + displacement [(d8, An)]
	eamodeFlagAregIndIndex   = eamodeFlag(1 << 6)  // Address register indirect + index [(d8, An, Xn)]
	eamodeFlagAbsW           = eamodeFlag(1 << 7)  // Absolute (16-bit sign-extended) [xxx.w]
	eamodeFlagAbsL           = eamodeFlag(1 << 8)  // Absolute (32-bit) [xxx.l]
	eamodeFlagPcIndDisp      = eamodeFlag(1 << 9)  // PC indirect + displacement [(d8, PC)]
	eamodeFlagPcIndIndex     = eamodeFlag(1 << 10) // PC indirect + index [(d8, PC)]
	eamodeFlagImm            = eamodeFlag(1 << 11) // Immediate (#xxx)

	// Data addressing modes
	eamodeFlagData = eamodeFlagDreg |
		eamodeFlagAregInd |
		eamodeFlagAregIndPostinc |
		eamodeFlagAregIndPredec |
		eamodeFlagAregIndDisp |
		eamodeFlagAregIndIndex |
		eamodeFlagPcIndDisp |
		eamodeFlagPcIndIndex |
		eamodeFlagAbsW |
		eamodeFlagAbsL |
		eamodeFlagImm

	// Memory addressing modes
	eamodeFlagMem = eamodeFlagAregInd |
		eamodeFlagAregIndPostinc |
		eamodeFlagAregIndPredec |
		eamodeFlagAregIndDisp |
		eamodeFlagAregIndIndex |
		eamodeFlagPcIndDisp |
		eamodeFlagPcIndIndex |
		eamodeFlagAbsW |
		eamodeFlagAbsL |
		eamodeFlagImm

	// Control addressing modes
	eamodeFlagCtrl = eamodeFlagAregInd |
		eamodeFlagAregIndDisp |
		eamodeFlagAregIndIndex |
		eamodeFlagPcIndDisp |
		eamodeFlagPcIndIndex |
		eamodeFlagAbsW |
		eamodeFlagAbsL

	// Alterable addressing modes
	eamodeFlagAlter = eamodeFlagDreg |
		eamodeFlagAreg |
		eamodeFlagAregInd |
		eamodeFlagAregIndPostinc |
		eamodeFlagAregIndPredec |
		eamodeFlagAregIndDisp |
		eamodeFlagAregIndIndex |
		eamodeFlagAbsW |
		eamodeFlagAbsL
)

func (eamode eamodeFlag) getConstNames() []string {
	res := make([]string, 0)
	if (eamode & eamodeFlagDreg) != 0 {
		res = append(res, "eamodeDreg")
	}
	if (eamode & eamodeFlagAreg) != 0 {
		res = append(res, "eamodeAreg")
	}
	if (eamode & eamodeFlagAregInd) != 0 {
		res = append(res, "eamodeAregInd")
	}
	if (eamode & eamodeFlagAregIndPostinc) != 0 {
		res = append(res, "eamodeAregIndPostinc")
	}
	if (eamode & eamodeFlagAregIndPredec) != 0 {
		res = append(res, "eamodeAregIndPredec")
	}
	if (eamode & eamodeFlagAregIndDisp) != 0 {
		res = append(res, "eamodeAregIndDisp")
	}
	if (eamode & eamodeFlagAregIndIndex) != 0 {
		res = append(res, "eamodeAregIndIndex")
	}
	if (eamode & eamodeFlagAbsW) != 0 {
		res = append(res, "eamodeAbsW")
	}
	if (eamode & eamodeFlagAbsL) != 0 {
		res = append(res, "eamodeAbsL")
	}
	if (eamode & eamodeFlagPcIndDisp) != 0 {
		res = append(res, "eamodePcIndDisp")
	}
	if (eamode & eamodeFlagPcIndIndex) != 0 {
		res = append(res, "eamodePcIndIndex")
	}
	if (eamode & eamodeFlagImm) != 0 {
		res = append(res, "eamodeImm")
	}
	return res
}

//==============================================================================
// 68000 Operation size
//==============================================================================

type opsize uint8

const (
	opsizeByte = opsize(iota)
	opsizeWord
	opsizeLong
)

//==============================================================================
// Source code output
//==============================================================================

var dest *bufio.Writer
var indent int

const (
	sourceIndent = 4
)

func writeStrings(strs []string) {
	for _, s := range strs {
		if _, err := dest.WriteString(s); err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
	}
}

func emitln(format string, args ...any) {
	writeStrings([]string{strings.Repeat(" ", indent), fmt.Sprintf(format, args...), "\n"})
}
func emitBeginBlock(format string, args ...any) {
	writeStrings([]string{strings.Repeat(" ", indent), fmt.Sprintf(format, args...), " {\n"})
	indent += sourceIndent
}
func emitEndBlock(format string, args ...any) {
	indent -= sourceIndent
	writeStrings([]string{strings.Repeat(" ", indent), "}", fmt.Sprintf(format, args...), "\n"})
}
func emitEndBeginBlock(format string, args ...any) {
	indent -= sourceIndent
	// FIXME: Add space beteen the code and }
	writeStrings([]string{strings.Repeat(" ", indent), "}", fmt.Sprintf(format, args...), " {\n"})
	indent += sourceIndent
}
