#!/usr/bin/env node
var os = require("os");
var fs = require("fs");
var path = require("path");

// 'aix', 'darwin', 'freebsd','linux', 'openbsd', 'sunos', and 'win32'
let platform = os.platform();

// 'arm', 'arm64', 'ia32', 'mips', 'mipsel', 'ppc', 'ppc64', 's390', 's390x', and 'x64'
const arch = os.arch();

// let suffix = ""
// if(platform == "win32") {
//     suffix = ".exe";
//     platform = "windows"
// }

const targetPackage = "@gitwho/" + platform + "-" + arch;

const targetLibPath = path.dirname(require.resolve('esbuild'))


console.log(`Using bin "${binName}" - TODO download package for platform`)

// Codes below were adapted from https://github.com/evanw/esbuild/blob/main/lib/npm/node-install.ts

// function installUsingNPM(pkg, subpath, binPath) {
//     // Erase "npm_config_global" so that "npm install --global esbuild" works.
//     // Otherwise this nested "npm install" will also be global, and the install
//     // will deadlock waiting for the global installation lock.
//     const env = { ...process.env, npm_config_global: undefined }
  
//     // Create a temporary directory inside the "esbuild" package with an empty
//     // "package.json" file. We'll use this to run "npm install" in.
//     const esbuildLibDir = path.dirname(require.resolve('esbuild'))
//     const installDir = path.join(esbuildLibDir, 'npm-install')
//     fs.mkdirSync(installDir)
//     try {
//       fs.writeFileSync(path.join(installDir, 'package.json'), '{}')
  
//       // Run "npm install" in the temporary directory which should download the
//       // desired package. Try to avoid unnecessary log output. This uses the "npm"
//       // command instead of a HTTP request so that it hopefully works in situations
//       // where HTTP requests are blocked but the "npm" command still works due to,
//       // for example, a custom configured npm registry and special firewall rules.
//       child_process.execSync(`npm install --loglevel=error --prefer-offline --no-audit --progress=false ${pkg}@${versionFromPackageJSON}`,
//         { cwd: installDir, stdio: 'pipe', env })
  
//       // Move the downloaded binary executable into place. The destination path
//       // is the same one that the JavaScript API code uses so it will be able to
//       // find the binary executable here later.
//       const installedBinPath = path.join(installDir, 'node_modules', pkg, subpath)
//       fs.renameSync(installedBinPath, binPath)
//     } finally {
//       // Try to clean up afterward so we don't unnecessarily waste file system
//       // space. Leaving nested "node_modules" directories can also be problematic
//       // for certain tools that scan over the file tree and expect it to have a
//       // certain structure.
//       try {
//         removeRecursive(installDir)
//       } catch {
//         // Removing a file or directory can randomly break on Windows, returning
//         // EBUSY for an arbitrary length of time. I think this happens when some
//         // other program has that file or directory open (e.g. an anti-virus
//         // program). This is fine on Unix because the OS just unlinks the entry
//         // but keeps the reference around until it's unused. There's nothing we
//         // can do in this case so we just leave the directory there.
//       }
//     }
//   }

//   function removeRecursive(dir) {
//     for (const entry of fs.readdirSync(dir)) {
//       const entryPath = path.join(dir, entry)
//       let stats
//       try {
//         stats = fs.lstatSync(entryPath)
//       } catch {
//         continue; // Guard against https://github.com/nodejs/node/issues/4760
//       }
//       if (stats.isDirectory()) removeRecursive(entryPath)
//       else fs.unlinkSync(entryPath)
//     }
//     fs.rmdirSync(dir)
//   }