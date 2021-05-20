#!/usr/bin/env node

const { spawnSync } = require('child_process');

if (process.argv.length < 3) {
  console.log(`Usage: ${process.argv[0]} ${process.argv[1]} pkg@version`);
  process.exit(1);
}

let result = spawnSync('npm', ['init', '--force'], { stdio: 'inherit' });
if (result.status != 0) {
  throw 'Failed to init npm';
}

const pkgAndVersion = process.argv[2];
result = spawnSync('npm', ['install', pkgAndVersion], { stdio: 'pipe', encoding: 'utf8' });
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

const pkg = pkgAndVersion.split('@')[0];
try {
  require(pkg);
} catch (e) {
  console.log(`Failed to import ${pkg}: ${e}`);
}
