#!/usr/bin/env node

// eslint no-var: 0
// jshint esversion: 6
// jshint node: true
'use strict';

const {spawnSync} = require('child_process');
const fs = require('fs');
const process = require('process');

const executionLogPath = '/execution.log';

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

function redirectConsoleWrite(stdoutWrite, stderrWrite) {
  process.stdout.write = stdoutWrite;
  process.stderr.write = stderrWrite;
}

function importPkg(pkg) {
  try {
    require(pkg.name);
  } catch (e) {
    console.log(`Failed to import ${pkg.name}: ${e}`);
  }
}

function executePkg(pkg) {
  // if we're here, module importing should have worked in import phase
  let mod = require(pkg.name);

  executeModule(pkg.name, mod);
}

function executeModule(name, mod) {
  // redirect stdout and stderr to execution log during execution phase
  let executionLogStream = null;
  try {
    executionLogStream = fs.createWriteStream(executionLogPath, {'flags': 'a'});
  } catch (e) {
    console.log(`Failed to open execution log: ${e}`);
    return;
  }

  const defaultStdoutWrite = process.stdout.write;
  const defaultStderrWrite = process.stderr.write;
  const executionLogWrite = function () {
    //defaultStdoutWrite.apply(process.stdout, arguments);
    executionLogStream.write.apply(executionLogStream, arguments);
  };

  redirectConsoleWrite(executionLogWrite, executionLogWrite);

  let executionErr = null;
  try {
    executeModuleCode(pkg.name, mod);
  } catch (e) {
    executionErr = e;
    console.log(`Error while executing ${pkg.name} module code: ${e}`);
  } finally {
    executionLogStream.close();
    // restore default console behaviour
    redirectConsoleWrite(defaultStdoutWrite, defaultStderrWrite);
  }

  if (executionErr !== null) {
    // log to normal console too
    console.log(`Error while executing ${pkg.name} module code: ${executionErr}`);
  }
}

// Adapted from https://stackoverflow.com/a/72326559.
// Premise is that classes have a non-writable prototype
function isES6Class(obj) {
  if (typeof obj !== 'function') {
    return false;
  }

  const descriptor = Object.getOwnPropertyDescriptor(obj, 'prototype');
  return descriptor && !descriptor.writable;
}

// Best-effort execution of as much code (functions, classes) of the module code as possible
function executeModuleCode(name, mod) {
  console.log(`[module] ${name}`);
  console.log(mod);

  // How to tell if something is a function vs class:
  // - node uses complex internal logic in util.inspect() function [1]
  // - StackOverflow thread discussing how complex this problem is [2]
  //
  // Solution uses code from an answer in [2]
  // [1] https://github.com/nodejs/node/blob/main/lib/internal/util/inspect.js
  // [2] https://stackoverflow.com/questions/30758961/how-to-check-if-a-variable-is-an-es6-class-declaration/72326559
  const callableNames = Object.keys(mod).filter((key) => typeof mod[key] === 'function');
  console.log(`[keys] ${callableNames}`);

  // Call all the exported names. If there is a TypeError, it's
  // probably because it is a class, so we'll try again using new.
  // NOTE: this is a best-effort approach and there are lots of ways
  // it can fail. In particular, function arguments are not yet supported
  // TODO basic support for function arguments

  // https://nodejs.org/api/process.html#event-uncaughtexception
  // tl;dr this may cause things to break, but we're in a sandbox so we'll do it anyway
  process.on('uncaughtException', (err, origin) => {
	  console.log('[uncaught exception]');
	  console.log(err);
	  console.log(origin);
  });

  // https://nodejs.org/api/process.html#event-unhandledrejection
  process.on('unhandledRejection', (reason, promise) => {
	  console.log('[unhandled rejection]');
	  console.log(reason);
	  console.log(promise);
  });

  for (const name of callableNames) {
    try {
      // TODO call each function or class in a separate thread or sandbox,
      //  so that functions that block can be interrupted
      if (isES6Class(mod[name])) {
        console.log(`[class] ${name}`);
        const instance = new mod[name]();
		// TODO call instance methods
      } else {
        console.log(`[function] ${name}`);
        mod[name]();
      }
    } catch (err) {
      console.log(`[error]: ${err}`);
    }
  }
}

const phases = new Map([
  ['all', [install, importPkg]],
  ['install', [install]],
  ['import', [importPkg]],
  ['execute', [executePkg]],
]);

const argv = process.argv;
const nodeBin = argv.shift();
const scriptPath = argv.shift();

if (argv.length < 2 || argv > 4) {
  console.log(`Usage: ${nodeBin} ${scriptPath} [--local file | --version version] phase pkg`);
  process.exit(1);
}

// Parse the arguments manually to avoid introducing unnecessary dependencies
// and side effects that add noise to the strace output.
let localFile = null;
let ver = null;
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
