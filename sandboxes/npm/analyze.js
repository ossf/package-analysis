#!/usr/bin/env node

const { spawnSync } = require('child_process');
const { argv } = require('process');

console.log(argv);
const nodeBin = argv.shift();
const scriptPath = argv.shift();
console.log(argv);

if (argv.length < 2 || argv > 4) {
  console.log(`Usage: ${nodeBin} ${scriptPath} [--local file | --version version] phase pkg`);
  process.exit(1);
}

var localFile = null;
var ver = null;
switch (argv[0]) {
  case '--local':
    argv.shift();
    localFile = argv.shift();
    break;
  case '--version':
    argv.shift();
    ver = argv.shift();
    break;
}

const phase = argv.shift();
const pkg = argv.shift();

if (phase != 'all') {
  console.log('Only "all" phase is supported at the moment.');
  process.exit(1);
}

// Specify the package to install.
const installPkg = localFile ? localFile : (ver ? `${pkg}@${ver}` : pkg);

let result = spawnSync('npm', ['init', '--force'], { stdio: 'inherit' });
if (result.status != 0) {
  throw 'Failed to init npm';
}

result = spawnSync('npm', ['install', installPkg], { stdio: 'pipe', encoding: 'utf8' });
console.log(result.stdout + result.stderr);

if (result.status == 0) {
  console.log('Install succeeded.');
} else {
  if (result.stderr.includes('is not in the npm registry.')) {
    process.exit(0);
  }
  console.log('Install failed.');
  process.exit(1);
}

try {
  require(pkg);
} catch (e) {
  console.log(`Failed to import ${pkg}: ${e}`);
}
