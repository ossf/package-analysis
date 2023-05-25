#!/usr/bin/env node

// eslint no-var: 0
// jshint esversion: 6
// jshint node: true
"use strict";

const {spawnSync} = require('child_process');
const {argv} = require('process');

function install(pkg) {
  // Specify the package to install.
  const installPkg = pkg.localFile ? pkg.localFile : (pkg.version ? `${pkg.name}@${pkg.version}` : pkg.name);

  let result = spawnSync('npm', ['init', '--force'], {stdio: 'inherit'});
  if (result.status !== 0) {
    throw 'Failed to init npm';
  }

  result = spawnSync('npm', ['install', installPkg], {stdio: 'inherit'});
  if (result.status === 0) {
    console.log('Install succeeded.');
  } else {
    // Always exit on failure.
    // Install failing is either an interesting issue, or an opportunity to
    // improve the analysis.
    console.log('Install failed.');
    process.exit(1);
  }
}

function importPkg(pkg) {
  try {
    const mod = require(pkg.name);
    useModule(mod);
  } catch (e) {
    console.log(`Failed to import ${pkg.name}: ${e}`);
  }
}

// Adapted from https://stackoverflow.com/a/72326559.
// Premise is that classes have a non-writable prototype
function isES6Class(obj) {
  if (typeof obj !== "function") {
    return false;
  }

  const descriptor = Object.getOwnPropertyDescriptor(obj, "prototype");

  return descriptor && !descriptor.writable;
}

// best-effort execution of code in the module
function useModule(modulePath) {
  console.log("[module] " + modulePath);

  //console.log(module);

  // How to tell if something is a function vs class:
  // - node uses complex internal logic in util.inspect() function [1]
  // - StackOverflow thread discussing how complex this problem is [2]
  //
  // Solution uses code from an answer in [2]

  // [1] https://github.com/nodejs/node/blob/main/lib/internal/util/inspect.js
  // [2] https://stackoverflow.com/questions/30758961/how-to-check-if-a-variable-is-an-es6-class-declaration/72326559
  const callableNames = Object.keys(module).filter(
    (key) => typeof module[key] === "function"
  );

  // Call all the exported names. If there is a
  // TypeError, it's probably because it is a class,
  // so we'll try again using new
  for (const name of callableNames) {
    try {
      if (isES6Class(module[name])) {
        console.log("[class] " + name);
        const instance = new module[name]();
      } else {
        console.log("[function] " + name);
        module[name]();
      }
    } catch (err) {
      console.log(err);
    }
  }
}

const phases = new Map([
  ['all', [install, importPkg]],
  ['install', [install]],
  ['import', [importPkg]]
]);

const nodeBin = argv.shift();
const scriptPath = argv.shift();

if (argv.length < 2 || argv > 4) {
  console.log(`Usage: ${nodeBin} ${scriptPath} [--local file | --version version] phase pkg`);
  process.exit(1);
}

// Parse the arguments manually to avoid introducing unnecessary dependencies
// and side effects that add noise to the strace output.
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
const pkgName = argv.shift();
const pkg = {
  name: pkgName,
  version: ver,
  localFile: localFile,
};

if (!Array.from(phases.keys()).includes(phase)) {
  console.log(`Unknown phase ${phase} specified.`);
  process.exit(1);
}

// Execute the phase
phases.get(phase).forEach((f) => f(pkg));