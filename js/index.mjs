import fs from 'node:fs/promises';
import { argv } from 'node:process';
import CPUClient from './cpu.mjs';
import path from 'node:path';

const TEST_TIME_LIMIT = 1; // Seconds
const servAddr = '127.0.0.1';
let testPath = 'm68000/v1/';
let testNameFilters = [];

for (let i = 2; i < argv.length; i++) {
    switch (argv[i]) {
        // Test directory/file
        case '-t':
            i++;
            if (argv[i] == undefined) {
                console.error(`Expected argument after -t`);
                process.exit(1);
            }
            testPath = argv[i];
            break;
        // Test name filter (-f)
        case '-f':
            i++;
            if (argv[i] == undefined) {
                console.error(`Expected argument after -f`);
                process.exit(1);
            }
            testNameFilters.push(argv[i]);
            break;
    }
}

const ram = new Uint8Array(16 * 1024 * 1024);
const cpu = new CPUClient();
let assertedRamAddr = 0;
let execLogs = [];

function hex(x) {
    // If the MSB is set, it will become negative number in JS.
    if (x < 0) {
        const lower = (x & 0xfffffff).toString(16);
        const upper = ((x >> 28) & 0xf).toString(16);
        return upper + lower;
    }
    return x.toString(16);
}

cpu.onTraceExec = (pc, ir, disasm) => {
    execLogs.push(` EXEC | pc=${hex(pc)} ir=${hex(ir)} ${disasm}`);
};
cpu.onTraceExc = (exc, pc, ir, errAddr, errFlags) => {
    let extInfo = '';
    if (ir !== undefined) {
        extInfo = `ir=${hex(ir)} errAddr=${hex(errAddr)} `;
        extInfo += `errFlags=${hex(errFlags)}`;
    }
    execLogs.push(` !EXC | exc=${hex(exc)} pc=${hex(pc)} ${extInfo}`);
};
cpu.onAddressAsserted = (addr) => {
    assertedRamAddr = addr;
    return true;
};
cpu.onResetAsserted = () => {
    execLogs.push(`RESET |`);
};
cpu.onBusWrite = (ds, val) => {
    execLogs.push(
        `  BUS | W addr=${hex(assertedRamAddr)} ds=${ds} val=${hex(val)}`
    );
    if ((ds & CPUClient.DS_UPPER) != 0) {
        ram[assertedRamAddr] = val >> 8;
    }
    if ((ds & CPUClient.DS_LOWER) != 0) {
        ram[assertedRamAddr + 1] = val;
    }
};
cpu.onBusRead = (ds) => {
    let val = 0;
    if ((ds & CPUClient.DS_UPPER) != 0) {
        val |= ram[assertedRamAddr] << 8;
    }
    if ((ds & CPUClient.DS_LOWER) != 0) {
        val |= ram[assertedRamAddr + 1];
    }
    execLogs.push(
        `  BUS | R addr=${hex(assertedRamAddr)} ds=${ds} val=${hex(val)}`
    );
    return val;
};

let passCount = 0;
let failCount = 0;
let failMismatchCount = 0;

// The PC in the test data is actually next prefetch address, and we have to subtract 4 from it to get actual instruction address.
// From README at https://github.com/SingleStepTests/m68000:
// "... It's "next prefetch address" so it's +4 from where the test starts executing."
const calcRealPc = (pc) => pc - 4;

// Converts all number values to signed 32-bit integers
const numberToInt = (o) => {
    for (const k in o) {
        if (typeof o[k] == 'number') {
            // Doing any bit-op will convert it to 32-bit integer
            o[k] = o[k] | 0;
        } else if (typeof o[k] == 'object') {
            numberToInt(o[k]);
        }
    }
};

async function runTestFile(filename, fileText) {
    const tests = JSON.parse(fileText);

    for (const test of tests) {
        numberToInt(test);

        execLogs = [];
        let filterOk = testNameFilters.length === 0;
        for (const f of testNameFilters) {
            if (test.name.indexOf(f) != -1) {
                filterOk = true;
            }
        }
        if (!filterOk) {
            continue;
        }
        const initial = test.initial;
        const final = test.final;
        const loadPromises = [];

        //------------------------------------------------------------------
        // Setup initial state
        //------------------------------------------------------------------
        // Load data registers(D0~D7) --------------------------------------
        loadPromises.push(
            cpu.writeDreg(0, initial.d0),
            cpu.writeDreg(1, initial.d1),
            cpu.writeDreg(2, initial.d2),
            cpu.writeDreg(3, initial.d3),
            cpu.writeDreg(4, initial.d4),
            cpu.writeDreg(5, initial.d5),
            cpu.writeDreg(6, initial.d6),
            cpu.writeDreg(7, initial.d7)
        );

        // Load address registers(A0~A6) -----------------------------------
        loadPromises.push(
            cpu.writeAreg(0, initial.a0),
            cpu.writeAreg(1, initial.a1),
            cpu.writeAreg(2, initial.a2),
            cpu.writeAreg(3, initial.a3),
            cpu.writeAreg(4, initial.a4),
            cpu.writeAreg(5, initial.a5),
            cpu.writeAreg(6, initial.a6)
        );

        // Load A7(SSP and USP), PC, and SR --------------------------------
        loadPromises.push(
            cpu.writeSsp(initial.ssp),
            cpu.writeUsp(initial.usp),
            cpu.writePc(calcRealPc(initial.pc)),
            cpu.writeSr(initial.sr)
        );

        // Clear destination RAM -------------------------------------------
        for (const [addr, _] of final.ram) {
            ram[addr] = 0;
        }

        // Load RAM contents -----------------------------------------------
        for (const [addr, val] of initial.ram) {
            ram[addr] = val;
        }

        //------------------------------------------------------------------
        // Run the CPU
        //------------------------------------------------------------------
        let failed = false;

        try {
            await Promise.all(loadPromises);
        } catch (e) {
            console.log('State load error! Skipping this test...');
            console.log(e);
            failCount++;
            continue;
        }
        await cpu.unstop();
        let runStartTime = new Date();
        const finalPc = calcRealPc(final.pc);
        while (true) {
            const elapsed = new Date() - runStartTime;
            if (TEST_TIME_LIMIT * 1000 <= elapsed) {
                console.log(
                    `>>> Execution timeout (Expected final PC: ${hex(finalPc)})`
                );
                failed = true;
                break;
            }
            try {
                await cpu.tick();
            } catch (e) {
                console.log('>>> CPU Error');
                console.log(e);
                failed = true;
                break;
            }
            const isStopped = await cpu.isStopped();
            if (isStopped) {
                break;
            }
            const pc = await cpu.readPc();
            if (pc === finalPc) {
                break;
            }
        }

        //------------------------------------------------------------------
        // Compare the final state
        //------------------------------------------------------------------

        let mismatchLog = [];
        const onMismatch = (regName, expect, got) => {
            mismatchLog.push(
                `${regName} mismatch: Expected ${hex(expect)}, Got ${hex(got)}`
            );
            failed = true;
        };
        const compareReg = async (regName, expect, readReg) => {
            expect = expect;
            const got = await readReg();
            if (expect !== got) {
                onMismatch(regName, expect, got);
            }
        };
        const comparePromises = [];

        // Compare data registers(D0~D7) -----------------------------------
        const compareDreg = async (reg, expect) => {
            await compareReg(
                `D${reg}`,
                expect,
                async () => await cpu.readDreg(reg)
            );
        };
        comparePromises.push(
            compareDreg(0, final.d0),
            compareDreg(1, final.d1),
            compareDreg(2, final.d2),
            compareDreg(3, final.d3),
            compareDreg(4, final.d4),
            compareDreg(5, final.d5),
            compareDreg(6, final.d6),
            compareDreg(7, final.d7)
        );

        // Compare address registers(A0~A6) --------------------------------
        const compareAreg = async (reg, expect) => {
            await compareReg(
                `A${reg}`,
                expect,
                async () => await cpu.readAreg(reg)
            );
        };
        comparePromises.push(
            compareAreg(0, final.a0),
            compareAreg(1, final.a1),
            compareAreg(2, final.a2),
            compareAreg(3, final.a3),
            compareAreg(4, final.a4),
            compareAreg(5, final.a5),
            compareAreg(6, final.a6)
        );

        // Compare A7(SSP and USP), PC, and SR -----------------------------

        // Some of flags in certain instructions are undefined in 68000, so we ignore those.
        // Unfortunately there are still some undefined cases that we can't test here.
        let ignoreFlags = 0;
        if (
            test.name.indexOf(' ABCD ') !== -1 ||
            test.name.indexOf(' SBCD ') !== -1 ||
            test.name.indexOf(' NBCD ') !== -1
        ) {
            ignoreFlags = CPUClient.CCR_FLAG_N | CPUClient.CCR_FLAG_Z;
        } else if (test.name.indexOf(' CHK ') !== -1) {
            ignoreFlags =
                CPUClient.CCR_FLAG_Z |
                CPUClient.CCR_FLAG_V |
                CPUClient.CCR_FLAG_C;
        } else if (test.name.indexOf(' DIVS ') !== -1) {
            ignoreFlags = CPUClient.CCR_FLAG_N | CPUClient.CCR_FLAG_Z;
        }

        comparePromises.push(
            compareReg('SSP', final.ssp, async () => await cpu.readSsp()),
            compareReg('USP', final.usp, async () => await cpu.readUsp()),
            compareReg(
                'PC',
                calcRealPc(final.pc),
                async () => await cpu.readPc()
            ),
            compareReg(
                'SR',
                final.sr & ~ignoreFlags,
                async () => (await cpu.readSr()) & ~ignoreFlags
            )
        );

        try {
            await Promise.all(comparePromises);
        } catch (e) {
            console.log('Compare error!');
            console.log(e);
            failed = true;
            break;
        }

        // Compare RAM contents --------------------------------------------
        for (const [addr, expect] of final.ram) {
            const got = ram[addr];
            if (got != expect) {
                onMismatch(`Memory at ${hex(addr)}`, expect, got);
            }
        }

        //------------------------------------------------------------------
        // Show the report
        //------------------------------------------------------------------

        if (failed) {
            console.log(`[${filename}] ${test.name}`);
            if (mismatchLog.length !== 0) {
                console.log(`${mismatchLog.length} mismatches found!`);
                for (const line of mismatchLog) {
                    console.log(`>>> ${line}`);
                }
                failMismatchCount++;
            }
            console.log('--- Execution log');
            for (const line of execLogs) {
                console.log(`>>> ${line}`);
            }
            failCount++;
        } else {
            passCount++;
        }
    }
}

let filepaths = [];
if ((await fs.stat(testPath)).isDirectory()) {
    filepaths = (await fs.readdir(testPath)).map((x) => path.join(testPath, x));
} else {
    filepaths = [testPath];
}

cpu.connect(servAddr, async () => {
    await cpu.setTraceExec(true);
    await cpu.setTraceExc(true);

    const beginTime = new Date();
    console.log('Connected to server');
    for (const file of filepaths) {
        if (!file.endsWith('.json')) {
            continue;
        }
        const fileText = (await fs.readFile(file)).toString();
        await runTestFile(file, fileText);
    }
    console.log(
        `TEST FINISHED: ${passCount} passed, ${failCount} failed(${failMismatchCount} mismatches)`
    );
    cpu.bye();

    const took = Math.floor((new Date() - beginTime) / 1000);
    const mins = Math.floor(took / 60);
    const secs = took % 60;
    console.log(`Tests took ${mins} minutes ${secs} seconds`);
});
