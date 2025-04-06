import net from 'node:net';

const SERVER_PORT = 6800;

export default class CPUClient {
    #client = undefined;
    #inboxBuf = [];
    #sentBytesSum = 0;
    #recvBytesSum = 0;
    #connStartTime = undefined;

    static DS_UPPER = 1 << 0; // UDS=1 LDS=0; Only upper 8-bit of 16-bit data bus is active
    static DS_LOWER = 1 << 1; // UDS=0 LDS=1; Only lower 8-bit of 16-bit data bus is active
    static DS_BOTH = CPUClient.DS_UPPER | CPUClient.DS_LOWER; // UDS=1 LDS=1; All of 16-bit data bus is active

    static CCR_FLAG_C = 1 << 0;
    static CCR_FLAG_V = 1 << 1;
    static CCR_FLAG_Z = 1 << 2;
    static CCR_FLAG_N = 1 << 3;
    static CCR_FLAG_X = 1 << 4;

    // Takes address, and returns boolean indicating whether a valid device is there or not.
    onAddressAsserted = (_addr) => {
        throw new Error('not implemented');
    };
    // Reads from the last asserted address, and returns the result.
    onBusRead = (_ds) => {
        throw new Error('not implemented');
    };
    // Writes to the last asserted address.
    onBusWrite = () => {
        throw new Error('not implemented');
    };
    // RESET signal was asserted
    onResetAsserted = () => {
        throw new Error('not implemented');
    };
    // Called if execution tracing is enabled
    onTraceExec = (_pc, _ir, _disasm) => {
        throw new Error('not implemented');
    };
    // Called if exception tracing is enabled
    // Parameters ir, errAddr, errFlags are not set if it's not a memory exception(i.e. Bus and Address error).
    onTraceExc = (_exc, _pc, _ir, _errAddr, _errFlags) => {
        throw new Error('not implemented');
    };

    constructor() {}

    // Queue for functions waiting for response
    #responseWaitQueue = [];

    connect(host, onConnected) {
        this.#client = net.createConnection({ host, port: SERVER_PORT }, () => {
            console.log('[CPUClient] Connected');
            onConnected();
            this.#connStartTime = new Date();
        });
        this.#client.on('end', () => {
            const secs = (new Date() - this.#connStartTime) / 1000;
            const sentStr = `${Math.floor(
                this.#sentBytesSum / secs
            )} bytes/sec`;
            const recvStr = `${Math.floor(
                this.#recvBytesSum / secs
            )} bytes/sec`;
            console.log(
                `[CPUClient] Disconnected - Sent ${sentStr}, Recv ${recvStr}`
            );
        });
        this.#client.on('data', (data) => {
            this.#recvBytesSum += data.length;
            for (let i = 0; i < data.length; i++) {
                this.#inboxBuf.push(data.readUint8(i));
            }
            while (true) {
                if (this.#inboxBuf.length === 0) {
                    break;
                }
                // Look at the first response byte to see what it means.
                // - If it's a message we can't understand, remove it and move to next one.
                // - If it's a message we can understand but need more data, exit the event handler.
                //   Then check it again next time we receive some more data.
                const tp = this.#inboxBuf[0];
                switch (tp) {
                    // ACK -----------------------------------------------------
                    case NETOP.ACK: {
                        if (this.#responseWaitQueue.length === 0) {
                            console.error(
                                'Got ACK but there are no requests...?'
                            );
                            this.#takeMsg('');
                            break;
                        }
                        const [callback, fmt] = this.#responseWaitQueue[0];
                        const res = this.#takeMsg(fmt);
                        if (res === undefined) {
                            // Try again next time
                            return;
                        }
                        callback('ok', res);
                        this.#responseWaitQueue.shift();
                        break;
                    }
                    // FAIL ----------------------------------------------------
                    case NETOP.FAIL: {
                        if (this.#responseWaitQueue.length === 0) {
                            console.error(
                                'Got FAIL but there are no requests...?'
                            );
                            this.#takeMsg('');
                            break;
                        }
                        const [callback, fmt] = this.#responseWaitQueue[0];
                        callback('error');
                        this.#responseWaitQueue.shift();
                        break;
                    }
                    // EVENT_ADDR_ASSERTED -------------------------------------
                    case NETOP.EVENT_ADDR_ASSERTED: {
                        const res = this.#takeMsg('l');
                        if (res === undefined) {
                            // Try again next time
                            return;
                        }
                        const [addr] = res;
                        if (!this.onAddressAsserted(addr)) {
                            this.#client.write(new Uint8Array([NETOP.FAIL]));
                        } else {
                            this.#client.write(new Uint8Array([NETOP.ACK]));
                        }
                        break;
                    }
                    // EVENT_READ_BUS ------------------------------------------
                    case NETOP.EVENT_READ_BUS: {
                        const res = this.#takeMsg('b');
                        if (res === undefined) {
                            // Try again next time
                            return;
                        }
                        const [ds] = res;
                        const val = this.onBusRead(ds);
                        this.#client.write(
                            new Uint8Array([NETOP.ACK, ...makeW(val)])
                        );
                        break;
                    }
                    // EVENT_WRITE_BUS -----------------------------------------
                    case NETOP.EVENT_WRITE_BUS: {
                        const res = this.#takeMsg('bw');
                        if (res === undefined) {
                            // Try again next time
                            return;
                        }
                        const [ds, val] = res;
                        this.onBusWrite(ds, val);
                        this.#client.write(new Uint8Array([NETOP.ACK]));
                        break;
                    }
                    // EVENT_RESET ---------------------------------------------
                    case NETOP.EVENT_RESET: {
                        const res = this.#takeMsg('');
                        if (res === undefined) {
                            // Try again next time
                            return;
                        }
                        this.onResetAsserted();
                        this.#client.write(new Uint8Array([NETOP.ACK]));
                        break;
                    }
                    // EVENT_TRACE_EXEC ----------------------------------------
                    case NETOP.EVENT_TRACE_EXEC: {
                        const res = this.#takeMsg('lws');
                        if (res === undefined) {
                            // Try again next time
                            return;
                        }
                        const [pc, ir, disasm] = res;
                        this.onTraceExec(pc, ir, disasm);
                        this.#client.write(new Uint8Array([NETOP.ACK]));
                        break;
                    }
                    // EVENT_TRACE_EXC -----------------------------------------
                    case NETOP.EVENT_TRACE_EXC: {
                        const res = this.#takeMsg('bl');
                        if (res === undefined) {
                            // Try again next time
                            return;
                        }
                        const [exc, pc] = res;
                        this.onTraceExc(exc, pc);
                        this.#client.write(new Uint8Array([NETOP.ACK]));
                        break;
                    }
                    // EVENT_TRACE_EXC_MEM -----------------------------------------
                    case NETOP.EVENT_TRACE_EXC_MEM: {
                        const res = this.#takeMsg('blwlb');
                        if (res === undefined) {
                            // Try again next time
                            return;
                        }
                        const [exc, pc, ir, errAddr, errFlags] = res;
                        this.onTraceExc(exc, pc, ir, errAddr, errFlags);
                        this.#client.write(new Uint8Array([NETOP.ACK]));
                        break;
                    }
                    default: {
                        throw Error(`Unrecognized opbyte ${tp.toString(16)}`);
                    }
                }
            }
        });
    }

    bye() {
        this.#client.end(new Uint8Array([NETOP.BYE]));
    }

    async unstop() {
        const cmd = [NETOP.UNSTOP];
        return this.#sendCmd(cmd, '');
    }

    async isStopped() {
        const cmd = [NETOP.IS_STOPPED];
        return (await this.#sendCmd(cmd, 'b')) === 1;
    }

    async setTraceExec(v) {
        const cmd = [v ? NETOP.TRACE_EXEC_ON : NETOP.TRACE_EXEC_OFF];
        return this.#sendCmd(cmd, '');
    }

    async setTraceExc(v) {
        const cmd = [v ? NETOP.TRACE_EXC_ON : NETOP.TRACE_EXC_OFF];
        return this.#sendCmd(cmd, '');
    }

    async tick() {
        const cmd = [NETOP.TICK];
        return this.#sendCmd(cmd, '');
    }

    async writeDreg(reg, val) {
        const cmd = [NETOP.WRITE_DREG, reg, ...makeL(val)];
        return this.#sendCmd(cmd, '');
    }

    async readDreg(reg) {
        const cmd = [NETOP.READ_DREG, reg];
        return (await this.#sendCmd(cmd, 'l'))[0];
    }

    async writeAreg(reg, val) {
        const cmd = [NETOP.WRITE_AREG, reg, ...makeL(val)];
        return this.#sendCmd(cmd, '');
    }

    async readAreg(reg) {
        const cmd = [NETOP.READ_AREG, reg];
        return (await this.#sendCmd(cmd, 'l'))[0];
    }

    async writeSsp(val) {
        const cmd = [NETOP.WRITE_SSP, ...makeL(val)];
        return this.#sendCmd(cmd, '');
    }

    async readSsp() {
        const cmd = [NETOP.READ_SSP];
        return (await this.#sendCmd(cmd, 'l'))[0];
    }

    async writeUsp(val) {
        const cmd = [NETOP.WRITE_USP, ...makeL(val)];
        return this.#sendCmd(cmd, '');
    }

    async readUsp() {
        const cmd = [NETOP.READ_USP];
        return (await this.#sendCmd(cmd, 'l'))[0];
    }

    async writePc(val) {
        const cmd = [NETOP.WRITE_PC, ...makeL(val)];
        return this.#sendCmd(cmd, '');
    }

    async readPc() {
        const cmd = [NETOP.READ_PC];
        return (await this.#sendCmd(cmd, 'l'))[0];
    }

    async writeSr(val) {
        const cmd = [NETOP.WRITE_SR, ...makeW(val)];
        return this.#sendCmd(cmd, '');
    }

    async readSr() {
        const cmd = [NETOP.READ_SR];
        return (await this.#sendCmd(cmd, 'w'))[0];
    }

    #sendCmd(cmd, fmt) {
        cmd.forEach((e) => {
            if (typeof e !== 'number') {
                throw TypeError(`${e} is not a number`);
            }
        });
        let cmdName;
        for (const k in NETOP) {
            if (NETOP[k] == cmd[0]) {
                cmdName = k;
            }
        }
        if (cmdName === undefined) {
            throw TypeError(`${e} is not a valid opbyte`);
        }
        this.#client.write(new Uint8Array(cmd));
        return new Promise((resolve, reject) => {
            let o = [
                (status, data) => {
                    clearTimeout(timeout);
                    if (status === 'ok') {
                        resolve(data);
                    } else {
                        reject(
                            new Error(
                                `Server returned FAIL response(Command: ${cmdName})`
                            )
                        );
                    }
                },
                fmt,
            ];
            let timeout = setTimeout(() => {
                o[0] = (status, data) => {
                    console.error(
                        `Response arrived too late(Command: ${cmdName}). status=${status}, data=${data}`
                    );
                };
                reject(new Error(`Response timeout(Command: ${cmdName})`));
            }, 1000);
            this.#responseWaitQueue.push(o);
        });
    }

    // Returns undefined if buffered data is not sufficient yet.
    #takeMsg(fmt) {
        // Check if we have enough data buffered.
        let needed_len = 1; // Length of type byte
        for (let i = 0; i < fmt.length; i++) {
            switch (fmt[i]) {
                case 'b':
                    needed_len += 1;
                    break;
                case 'w':
                    needed_len += 2;
                    break;
                case 'l':
                    needed_len += 4;
                    break;
                case 's': {
                    if (this.#inboxBuf.length < 2) {
                        return;
                    }
                    const len = this.#inboxBuf[1];
                    needed_len += 1 + len;
                    break;
                }
                default:
                    console.error(`Unrecognized format char ${fmt[i]}`);
                    break;
            }
        }
        if (this.#inboxBuf.length < needed_len) {
            return undefined;
        }

        // Remove the type byte
        this.#inboxBuf.shift();
        // Parse the result
        let results = [];
        for (let i = 0; i < fmt.length; i++) {
            switch (fmt[i]) {
                case 'b':
                    results.push(this.#inboxBuf.shift());
                    break;
                case 'w': {
                    const bytes = [
                        this.#inboxBuf.shift(),
                        this.#inboxBuf.shift(),
                    ];
                    results.push((bytes[0] << 8) | bytes[1]);
                    break;
                }
                case 'l': {
                    const bytes = [
                        this.#inboxBuf.shift(),
                        this.#inboxBuf.shift(),
                        this.#inboxBuf.shift(),
                        this.#inboxBuf.shift(),
                    ];
                    results.push(
                        (bytes[0] << 24) |
                            (bytes[1] << 16) |
                            (bytes[2] << 8) |
                            bytes[3]
                    );
                    break;
                }
                case 's': {
                    const len = this.#inboxBuf.shift();
                    const bytes = new Uint8Array(len);
                    for (let i = 0; i < len; i++) {
                        bytes[i] = this.#inboxBuf.shift();
                    }
                    const tdec = new TextDecoder('utf-8');
                    results.push(tdec.decode(bytes));
                    break;
                }
                default:
                    console.error(`Unrecognized format char ${fmt[i]}`);
                    break;
            }
        }
        return results;
    }
}

function makeW(v) {
    return [(v >> 8) & 0xff, v];
}
function makeL(v) {
    return [(v >> 24) & 0xff, (v >> 16) & 0xff, (v >> 8) & 0xff, v & 0xff];
}

const NETOP = {
    ACK: 0x00,
    FAIL: 0x01,

    BYE: 0x10,
    UNSTOP: 0x11,
    IS_STOPPED: 0x12,
    TRACE_EXEC_ON: 0x13,
    TRACE_EXEC_OFF: 0x14,
    TRACE_EXC_ON: 0x15,
    TRACE_EXC_OFF: 0x16,
    TICK: 0x1f,

    WRITE_DREG: 0x20,
    READ_DREG: 0x21,
    WRITE_AREG: 0x22,
    READ_AREG: 0x23,
    WRITE_SSP: 0x24,
    READ_SSP: 0x25,
    WRITE_USP: 0x26,
    READ_USP: 0x27,
    WRITE_PC: 0x28,
    READ_PC: 0x29,
    WRITE_SR: 0x2a,
    READ_SR: 0x2b,

    EVENT_ADDR_ASSERTED: 0x80,
    EVENT_READ_BUS: 0x81,
    EVENT_WRITE_BUS: 0x82,
    EVENT_RESET: 0x83,
    EVENT_TRACE_EXEC: 0x84,
    EVENT_TRACE_EXC: 0x85,
    EVENT_TRACE_EXC_MEM: 0x86,
};
