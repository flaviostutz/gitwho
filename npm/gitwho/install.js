#!/usr/bin/env node
var os = require("os");

// 'aix', 'darwin', 'freebsd','linux', 'openbsd', 'sunos', and 'win32'
const platform = os.platform();

// 'arm', 'arm64', 'ia32', 'mips', 'mipsel', 'ppc', 'ppc64', 's390', 's390x', and 'x64'
const arch = os.arch();

let suffix = ""
if(platform == "win32") {
    suffix = ".exe";
}

const binName = "gitwho-" + platform + "-" + arch + suffix;

console.log(`Using bin "${binname}" - TODO download package for platform`)
